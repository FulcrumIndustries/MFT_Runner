import os
import threading
from pyftpdlib.authorizers import DummyAuthorizer
from pyftpdlib.handlers import FTPHandler
from pyftpdlib.servers import FTPServer
from concurrent.futures import ThreadPoolExecutor

# Configuration
FTP_USER = "ftp"
FTP_PASS = "ftp"
FTP_PORT = 2121

# Get script directory and create receive folder
SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
RECEIVE_DIR = os.path.join(SCRIPT_DIR, "ftpreceive")
os.makedirs(RECEIVE_DIR, exist_ok=True)

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

if __name__ == "__main__":
    print("🚀 Starting FTP test server...")
    threading.Thread(target=start_ftp_server, daemon=True).start()
    
    # Keep main thread alive
    while True:
        pass
