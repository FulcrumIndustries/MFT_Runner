import os
import io
import threading
from pyftpdlib.authorizers import DummyAuthorizer
from pyftpdlib.handlers import FTPHandler
from pyftpdlib.servers import FTPServer
from concurrent.futures import ThreadPoolExecutor
from cryptography.hazmat.backends import default_backend
from cryptography.hazmat.primitives import serialization
from cryptography.hazmat.primitives import hashes
from cryptography.hazmat.primitives.asymmetric import rsa
from cryptography import x509
from cryptography.x509.oid import NameOID
from datetime import datetime, timedelta
import paramiko
from paramiko import Transport, ServerInterface, SFTPServerInterface, SFTPServer, SFTPAttributes, SFTPHandle
import socket
import errno
import sys
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
import time

# Configuration
FTP_USER = "ftp"
FTP_PASS = "ftp"
FTP_PORT = 2121
SFTP_PORT = 2222
HTTP_PORT = 8080

# Get script directory and create receive folder
SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
RECEIVE_DIR = os.path.join(SCRIPT_DIR, "ftpreceive")
os.makedirs(RECEIVE_DIR, exist_ok=True)

# Add a helper function to extract bytes (Not directly used here, but kept)
def extract_bytes(obj):
    if isinstance(obj, io.BytesIO):
        return obj.getvalue()
    if hasattr(obj, 'read'):
        try:
            data = obj.read()
            if isinstance(data, io.BytesIO):
                return data.getvalue()
            return data.encode() if not isinstance(data, bytes) else data
        except Exception as e:
            print("Error reading object:", e)
            return b''
    try:
        return bytes(obj)
    except:
        print(f"Cannot convert object of type {type(obj)} to bytes")
        return b""

def start_ftp_server():
    authorizer = DummyAuthorizer()
    authorizer.add_user(FTP_USER, FTP_PASS, RECEIVE_DIR, perm="elradfmwMT")
    handler = FTPHandler
    handler.authorizer = authorizer
    handler.banner = "MFT Test FTP Server Ready"
    handler.max_cons = 50
    handler.max_cons_per_ip = 50
    handler.timeout = 300
    handler.use_sendfile = sys.platform != "win32"  # Disable on Windows
    handler.tcp_no_delay = True
    handler.thread_pool = ThreadPoolExecutor(max_workers=50)
    server = FTPServer(("127.0.0.1", FTP_PORT), handler)
    server.max_cons = 50
    server.max_cons_per_ip = 50
    print(f"📁 FTP Server listening on port {FTP_PORT}")
    print(f"📂 Saving files to: {RECEIVE_DIR}")
    print(f"🔑 Credentials: {FTP_USER}/{FTP_PASS}")
    server.serve_forever()

def generate_certificate(prefix, common_name):
    key_path = os.path.join("AS2ServerCerts", f"{prefix}_key.pem")
    cert_path = os.path.join("AS2ServerCerts", f"{prefix}_cert.pem")
    os.makedirs(os.path.dirname(key_path), exist_ok=True)
    os.makedirs(os.path.dirname(cert_path), exist_ok=True)
    key = rsa.generate_private_key(public_exponent=65537, key_size=2048, backend=default_backend())
    subject = issuer = x509.Name([x509.NameAttribute(NameOID.COMMON_NAME, common_name)])
    cert = (x509.CertificateBuilder().subject_name(subject).issuer_name(issuer)
            .public_key(key.public_key()).serial_number(x509.random_serial_number())
            .not_valid_before(datetime.utcnow()).not_valid_after(datetime.utcnow() + timedelta(days=365))
            .add_extension(x509.KeyUsage(digital_signature=True, key_encipherment=True,
                                       content_commitment=False, data_encipherment=False,
                                       key_agreement=False, key_cert_sign=False, crl_sign=False,
                                       encipher_only=False, decipher_only=False), critical=True)
            .add_extension(x509.ExtendedKeyUsage([x509.ExtendedKeyUsageOID.SERVER_AUTH]), critical=False)
            .add_extension(x509.BasicConstraints(ca=True, path_length=None), critical=True)
            .sign(key, hashes.SHA256(), default_backend()))
    with open(key_path, "wb") as f:
        f.write(key.private_bytes(encoding=serialization.Encoding.PEM,
                                  format=serialization.PrivateFormat.TraditionalOpenSSL,
                                  encryption_algorithm=serialization.NoEncryption()))
    with open(cert_path, "wb") as f:
        f.write(cert.public_bytes(serialization.Encoding.PEM))
    print(f"🔐 Generated {prefix} certificate: {cert_path}")

class StubServer(ServerInterface):
    def check_auth_password(self, username, password):
        return paramiko.AUTH_SUCCESSFUL if username == FTP_USER and password == FTP_PASS else paramiko.AUTH_FAILED

    def check_channel_request(self, kind, chanid):
        return paramiko.OPEN_SUCCEEDED if kind == "session" else paramiko.OPEN_FAILED_ADMINISTRATIVELY_PROHIBITED

class StubSFTPHandle(SFTPHandle):
    def __init__(self, file_obj):
        super().__init__()
        self.file_obj = file_obj
        self._lock = threading.Lock()  # File access lock

    def close(self):
        with self._lock:  # Acquire lock for closing
            print(f"🔒 Closing file: {self.file_obj.name}")
            return self.file_obj.close()

    def read(self, offset, length):
        with self._lock:  # Acquire lock for reading
            print(f"📖 Reading {length} bytes from offset {offset} in {self.file_obj.name}")
            self.file_obj.seek(offset)
            return self.file_obj.read(length)

    def write(self, offset, data):
        with self._lock:  # Acquire lock for writing
            print(f"📝 Writing {len(data)} bytes at offset {offset} in {self.file_obj.name}")
            try:
                self.file_obj.seek(offset)
                self.file_obj.write(data)
                self.file_obj.flush()
                return paramiko.SFTP_OK
            except Exception as e:
                print(f"Write error: {e}")
                return paramiko.SFTP_FAILURE  # Use SFTP_FAILURE

    def stat(self):
        return self._get_attributes()

    def _get_attributes(self):
        with self._lock:
            try:
                stat = SFTPAttributes.from_stat(os.fstat(self.file_obj.fileno()))
                return stat
            except OSError as e:
                print(f"Error in stat: {e}")
                return paramiko.SFTPServer.convert_errno(e.errno)


class StubSFTPServer(SFTPServerInterface):
    def __init__(self, server):
        self.server = server

    def _realpath(self, path):
        return os.path.normpath(os.path.join(RECEIVE_DIR, path.lstrip('/')))

    def open(self, path, flags, attr):
        full_path = self._realpath(path)
        print(f"📝 Opening file: {full_path} (flags={flags:08x})")  # Log flags in hex

        try:
            if flags & os.O_CREAT:
                os.makedirs(os.path.dirname(full_path), exist_ok=True)
                if flags & os.O_APPEND:
                    mode = 'ab'
                elif flags & os.O_TRUNC:
                    mode = 'wb'
                else:
                    mode = 'wb'  # Create and write (overwrite if exists)
            elif flags & (os.O_WRONLY | os.O_RDWR):
                if flags & os.O_APPEND:
                  mode = 'ab'
                else:
                  mode = 'r+b'  # Open for reading and writing
            else:  # os.O_RDONLY or others
                mode = 'rb'
            print(mode)
            f = open(full_path, mode)
            handle = StubSFTPHandle(f)
            return handle
        except OSError as e:
            print(f"Error opening file: {e}")
            return paramiko.SFTPServer.convert_errno(e.errno)
    
    def _list_folder_helper(self, full_path):
        entries = []
        for item in os.listdir(full_path):
            item_path = os.path.join(full_path, item)
            attr = SFTPAttributes.from_stat(os.stat(item_path))
            attr.filename = item
            entries.append(attr)
        return entries

    def list_folder(self, path):
        full_path = self._realpath(path)
        print(f"📂 Listing folder: {full_path}")
        try:
            return self._list_folder_helper(full_path)
        except OSError as e:
            print(f"OSError listing folder: {e}")
            return paramiko.SFTPServer.convert_errno(e.errno)

    def stat(self, path):
        full_path = self._realpath(path)
        print(f"📊 Stat for: {full_path}")
        try:
            return SFTPAttributes.from_stat(os.stat(full_path))
        except OSError as e:
            print(f"OSError in stat: {e}")
            return paramiko.SFTPServer.convert_errno(e.errno)

    def lstat(self, path):
        full_path = self._realpath(path)
        print(f"📊 lstat for: {full_path}")
        try:
            return SFTPAttributes.from_stat(os.lstat(full_path))
        except OSError as e:
            print(f"OSError in lstat: {e}")
            return paramiko.SFTPServer.convert_errno(e.errno)

    def remove(self, path):
        full_path = self._realpath(path)
        print(f"🗑️ Removing file: {full_path}")
        try:
            os.remove(full_path)
            return paramiko.SFTP_OK
        except OSError as e:
            print(f"Error removing file: {e}")
            return paramiko.SFTPServer.convert_errno(e.errno)

    def rename(self, oldpath, newpath):
        old_full_path = self._realpath(oldpath)
        new_full_path = self._realpath(newpath)
        print(f"🔄 Renaming: {old_full_path} to {new_full_path}")
        try:
            os.rename(old_full_path, new_full_path)
            return paramiko.SFTP_OK
        except OSError as e:
            print(f"Error renaming: {e}")
            return paramiko.SFTPServer.convert_errno(e.errno)
            
    def mkdir(self, path, attr):
        full_path = self._realpath(path)
        print(f"Creating directory: {full_path}")
        try:
            os.makedirs(full_path, exist_ok=True)
            return paramiko.SFTP_OK
        except OSError as e:
            print(f"Error creating directory: {e}")
            return paramiko.SFTPServer.convert_errno(e.errno)

    def rmdir(self, path):
        full_path = self._realpath(path)
        print(f"🗑️ Removing directory: {full_path}")
        try:
            os.rmdir(full_path)
            return paramiko.SFTP_OK
        except OSError as e:
            print(f"Error removing directory: {e}")
            return paramiko.SFTPServer.convert_errno(e.errno)

    def chattr(self, path, attr):
        full_path = self._realpath(path)
        print(f"Changing attributes for: {full_path}")
        try:
            if attr.st_mode is not None:
                os.chmod(full_path, attr.st_mode)
            if (attr.st_uid is not None) and (attr.st_gid is not None):
                os.chown(full_path, attr.st_uid, attr.st_gid)
            return paramiko.SFTP_OK
        except OSError as e:
            print(f"Error changing attributes: {e}")
            return paramiko.SFTPServer.convert_errno(e.errno)
            
    def symlink(self, target_path, path):
         return paramiko.SFTPOperationNotSupported

    def readlink(self, path):
        return paramiko.SFTPOperationNotSupported

def start_sftp_server():
    host_key = paramiko.RSAKey.generate(2048)
    sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    sock.bind(('127.0.0.1', SFTP_PORT))
    sock.listen(50)
    print(f"📁 SFTP Server listening on port {SFTP_PORT}")
    print(f"📂 Saving files to: {RECEIVE_DIR}")
    print(f"🔑 Credentials: {FTP_USER}/{FTP_PASS}")
    with ThreadPoolExecutor(max_workers=50) as executor:
        while True:
            try:
                client, addr = sock.accept()
                print(f"Accepted connection from {addr}")
                client.settimeout(30)
                executor.submit(handle_sftp_connection, client, addr, host_key)
            except Exception as e:
                print(f"Error accepting connection: {e}")

def handle_sftp_connection(client, addr, host_key):
    try:
        transport = Transport(client)
        transport.add_server_key(host_key)
        transport.set_subsystem_handler('sftp', SFTPServer, StubSFTPServer)
        server = StubServer()
        transport.start_server(server=server)
        chan = transport.accept()
        if chan is None:
            print("No channel.")
            transport.close()
            return
        while transport.is_active():
            transport.join(1)
    except Exception as e:
        print(f"SFTP connection error: {e}")
    finally:
        transport.close()
        print(f"Connection from {addr} closed.")

class HTTPRequestHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        self.send_response(200)
        self.send_header("Content-type", "text/plain")
        self.end_headers()
        self.wfile.write(b"MFT Test HTTP Server Ready - POST files to /upload")
    
    def do_POST(self):
        try:
            # Handle missing Content-Length header
            if 'Content-Length' not in self.headers:
                self.send_error(411, "Length Required")
                return
                
            # Handle invalid Content-Length values
            try:
                content_length = int(self.headers['Content-Length'])
            except ValueError:
                self.send_error(400, "Invalid Content-Length")
                return
                
            file_data = self.rfile.read(content_length)
            # Get filename from Content-Disposition header or generate timestamped name
            content_disp = self.headers.get('Content-Disposition', '')
            if 'filename=' in content_disp:
                filename = content_disp.split('filename=')[1].strip('"')
            else:
                filename = f"http_upload_{int(time.time()*1000)}.dat"
            full_path = os.path.join(RECEIVE_DIR, filename)
            
            with open(full_path, "wb") as f:
                f.write(file_data)
            
            self.send_response(200)
            self.send_header("Content-type", "text/plain")
            self.end_headers()
            self.wfile.write(f"File saved as {full_path}".encode())
            print(f"🌐 HTTP: Received {len(file_data)} bytes saved as {full_path}")
            
        except Exception as e:
            self.send_error(500, str(e))
            print(f"HTTP Error: {e}")

def start_http_server():
    server = ThreadingHTTPServer(('127.0.0.1', HTTP_PORT), HTTPRequestHandler)
    print(f"🌐 HTTP Server listening on port {HTTP_PORT}")
    print(f"📂 Saving files to: {RECEIVE_DIR}")
    server.serve_forever()

if __name__ == "__main__":
    print("🚀 Starting FTP/SFTP/HTTP test servers...")
    threading.Thread(target=start_ftp_server, daemon=True).start()
    threading.Thread(target=start_sftp_server, daemon=True).start()
    threading.Thread(target=start_http_server, daemon=True).start()
    while True:
        pass