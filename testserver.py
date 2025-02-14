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

# Configuration
FTP_USER = "ftp"
FTP_PASS = "ftp"
FTP_PORT = 2121
SFTP_PORT = 2222

# Get script directory and create receive folder
SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
RECEIVE_DIR = os.path.join(SCRIPT_DIR, "ftpreceive")
os.makedirs(RECEIVE_DIR, exist_ok=True)

# Add a helper function to extract bytes from a part (Not directly used here, but kept for potential future use)
def extract_bytes(obj):
    # If the object is already a BytesIO, return its content
    if isinstance(obj, io.BytesIO):
        return obj.getvalue()
    if hasattr(obj, 'read'):
        try:
            # Try reading from the object
            data = obj.read()
            if isinstance(data, io.BytesIO):
                return data.getvalue()
            if not isinstance(data, bytes):
                return data.encode() # or str(data).encode(),  consider how you want to handle non-string data.
            return data

        except Exception as e:
            print("Error reading object:", e)
            return b''
    else:
        try: # attempt direct conversion if not readable
            return bytes(obj)
        except:
            print(f"Cannot convert object of type {type(obj)} to bytes")
            return b""



def start_ftp_server():
    # Set up user permissions (full access)
    authorizer = DummyAuthorizer()
    authorizer.add_user(
        username=FTP_USER,
        password=FTP_PASS,
        homedir=RECEIVE_DIR,
        perm="elradfmwMT"  # Add directory creation and file write permissions
    )

    # Configure server
    handler = FTPHandler
    handler.authorizer = authorizer
    handler.banner = "MFT Test FTP Server Ready"
    
    # Add these performance configurations
    handler.max_cons = 50  # Maximum simultaneous connections
    handler.max_cons_per_ip = 30  # Maximum connections from single IP
    handler.timeout = 300  # Connection timeout in seconds
    handler.use_sendfile = True  # Use efficient file transfer
    handler.tcp_no_delay = True  # Reduce latency

    # Configure thread pool executor
    handler.thread_pool = ThreadPoolExecutor(max_workers=30)

    # Bind to localhost only
    server = FTPServer(("127.0.0.1", FTP_PORT), handler)
    server.max_cons = 50  # Match handler setting
    server.max_cons_per_ip = 50
    
    print(f"📁 FTP Server listening on port {FTP_PORT}")
    print(f"📂 Saving files to: {RECEIVE_DIR}")
    print(f"🔑 Credentials: {FTP_USER}/{FTP_PASS}")
    server.serve_forever()



def generate_certificate(prefix, common_name):
    key_path = os.path.join("AS2ServerCerts", f"{prefix}_key.pem")
    cert_path = os.path.join("AS2ServerCerts", f"{prefix}_cert.pem")
    
    # Ensure directory exists
    os.makedirs(os.path.dirname(key_path), exist_ok=True)
    os.makedirs(os.path.dirname(cert_path), exist_ok=True)
    
    # Generate private key
    key = rsa.generate_private_key(
        public_exponent=65537,
        key_size=2048,
        backend=default_backend()
    )
    
    # Create self-signed certificate
    subject = issuer = x509.Name([
        x509.NameAttribute(NameOID.COMMON_NAME, common_name),
    ])
    
    cert = (
        x509.CertificateBuilder()
        .subject_name(subject)
        .issuer_name(issuer)
        .public_key(key.public_key())
        .serial_number(x509.random_serial_number())
        .not_valid_before(datetime.utcnow())
        .not_valid_after(datetime.utcnow() + timedelta(days=365))
        .add_extension(
            x509.KeyUsage(
                digital_signature=True,
                key_encipherment=True,
                content_commitment=False,
                data_encipherment=False,
                key_agreement=False,
                key_cert_sign=False,
                crl_sign=False,
                encipher_only=False,
                decipher_only=False
            ),
            critical=True
        )
        .add_extension(
            x509.ExtendedKeyUsage([x509.ExtendedKeyUsageOID.SERVER_AUTH]),
            critical=False
        )
        .add_extension(
            x509.BasicConstraints(ca=True, path_length=None),
            critical=True,
        )
        .sign(key, hashes.SHA256(), default_backend())
    )
    
    # Write private key to file
    with open(key_path, "wb") as f:
        f.write(key.private_bytes(
            encoding=serialization.Encoding.PEM,
            format=serialization.PrivateFormat.TraditionalOpenSSL,
            encryption_algorithm=serialization.NoEncryption(),
        ))
    
    # Write certificate to file
    with open(cert_path, "wb") as f:
        f.write(cert.public_bytes(serialization.Encoding.PEM))

    print(f"🔐 Generated {prefix} certificate: {cert_path}")

class StubServer(ServerInterface):
    def check_auth_password(self, username, password):
        if username == FTP_USER and password == FTP_PASS:
            return paramiko.AUTH_SUCCESSFUL
        return paramiko.AUTH_FAILED

    def check_channel_request(self, kind, chanid):
        if kind == "session":
            return paramiko.OPEN_SUCCEEDED
        return paramiko.OPEN_FAILED_ADMINISTRATIVELY_PROHIBITED

class StubSFTPHandle(paramiko.SFTPHandle):
    def __init__(self, file_obj):
        super().__init__()
        self.file_obj = file_obj

    def close(self):
        self.file_obj.close()
        return paramiko.SFTP_OK

    def write(self, offset, data):
        self.file_obj.seek(offset)
        self.file_obj.write(data)
        self.file_obj.flush()
        return paramiko.SFTP_OK

    def read(self, offset, length):
        self.file_obj.seek(offset)
        return self.file_obj.read(length)

    def stat(self):
        try:
            return SFTPAttributes.from_stat(os.fstat(self.file_obj.fileno()))
        except OSError as e:
            return paramiko.SFTPServer.convert_errno(e.errno)
            

class StubSFTPServer(SFTPServerInterface):
    def __init__(self, server, *largs, **kwargs):
        super(StubSFTPServer, self).__init__(server, *largs, **kwargs)
        self.root_path = RECEIVE_DIR  # Set the root path for SFTP operations


    def _realpath(self, path):
        """Convert path to absolute path, ensuring it's within RECEIVE_DIR."""
        normalized_path = os.path.normpath(os.path.join(self.root_path, path.lstrip('/')))
        if not normalized_path.startswith(self.root_path):
             raise Exception("Access denied: Attempt to access outside root directory.")
        return normalized_path

    def list_folder(self, path):
        """List files and directories within a given path."""
        full_path = self._realpath(path)
        print(f"Listing folder: {full_path}")  # Debugging
        try:
            entries = []
            for item in os.listdir(full_path):
                item_path = os.path.join(full_path, item)
                attr = SFTPAttributes.from_stat(os.stat(item_path))
                attr.filename = item
                entries.append(attr)
            return entries
        except OSError as e:
            print(f"OSError listing folder: {e}")  # More specific error logging
            return paramiko.SFTPServer.convert_errno(e.errno)


    def stat(self, path):
        """Return file/directory attributes."""
        full_path = self._realpath(path)
        print(f"Stat for: {full_path}")  # Debugging
        try:
            return SFTPAttributes.from_stat(os.stat(full_path))
        except OSError as e:
            print(f"OSError in stat: {e}")  # More specific error logging
            return paramiko.SFTPServer.convert_errno(e.errno)

    def lstat(self, path):
        """Return file/directory attributes (like stat, but doesn't follow symlinks)."""
        full_path = self._realpath(path)
        print(f"lstat for: {full_path}")
        try:
            return SFTPAttributes.from_stat(os.lstat(full_path))  # Use lstat
        except OSError as e:
            print(f"OSError in lstat: {e}")
            return paramiko.SFTPServer.convert_errno(e.errno)

    def open(self, path, flags, attr):
        """Open a file for read/write/append operations."""
        full_path = self._realpath(path)
        print(f"📝 Opening file: {full_path} (flags={flags})")
        os.makedirs(os.path.dirname(full_path), exist_ok=True)
        
        # Convert flags to proper mode
        mode = 'r+b' if (flags & os.O_WRONLY) else 'rb'  # Read mode for downloads
        
        try:
            f = open(full_path, mode)
            handle = StubSFTPHandle(f)
            handle.readfile = f  # Separate read file descriptor
            return handle
        except OSError as e:
            return paramiko.SFTPServer.convert_errno(e.errno)

    def remove(self, path):
        """Remove a file."""
        full_path = self._realpath(path)
        try:
            os.remove(full_path)
            return paramiko.SFTP_OK
        except OSError as e:
            return paramiko.SFTPServer.convert_errno(e.errno)

    def rename(self, oldpath, newpath):
        """Rename a file or directory."""
        old_full_path = self._realpath(oldpath)
        new_full_path = self._realpath(newpath)
        try:
            os.rename(old_full_path, new_full_path)
            return paramiko.SFTP_OK
        except OSError as e:
            return paramiko.SFTPServer.convert_errno(e.errno)

    def mkdir(self, path, attr):
        """Create a directory."""
        full_path = self._realpath(path)
        try:
            os.makedirs(full_path, exist_ok=True)  # exist_ok=True prevents error if dir exists
            return paramiko.SFTP_OK
        except OSError as e:
            return paramiko.SFTPServer.convert_errno(e.errno)

    def rmdir(self, path):
        """Remove a directory."""
        full_path = self._realpath(path)
        try:
            os.rmdir(full_path)  # rmdir only removes empty directories
            return paramiko.SFTP_OK
        except OSError as e:
            return paramiko.SFTPServer.convert_errno(e.errno)


    def chattr(self, path, attr):
        """ Change file/directory attributes.  (chmod/chown)"""
        full_path = self._realpath(path)

        try:
            if attr.st_mode is not None:
                os.chmod(full_path, attr.st_mode)
            if (attr.st_uid is not None) and (attr.st_gid is not None):
                os.chown(full_path, attr.st_uid, attr.st_gid)
            return paramiko.SFTP_OK
        except OSError as e:
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
    sock.listen(50)  # Listen with a backlog of 50

    print(f"📁 SFTP Server listening on port {SFTP_PORT}")
    print(f"📂 Saving files to: {RECEIVE_DIR}")
    print(f"🔑 Credentials: {FTP_USER}/{FTP_PASS}")

    # Use a ThreadPoolExecutor to handle multiple connections
    with ThreadPoolExecutor(max_workers=50) as executor:
        while True:
            try:
                client, addr = sock.accept()
                print(f"Accepted connection from {addr}")
                client.settimeout(30)
                executor.submit(handle_sftp_connection, client, addr, host_key)  # Submit the connection to the executor
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


if __name__ == "__main__":
    print("🚀 Starting FTP/SFTP test servers...")
    threading.Thread(target=start_ftp_server, daemon=True).start()
    threading.Thread(target=start_sftp_server, daemon=True).start()

    # Keep main thread alive
    while True:
        pass