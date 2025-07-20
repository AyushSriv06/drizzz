# 🔗 DrizLink - P2P File Sharing Application 🔗

A peer-to-peer file sharing application with integrated chat functionality, allowing users to connect, communicate, and share files directly with each other.

## ✨ Features

- **👤 User Authentication**: Connect with a username and maintain persistent sessions
- **🔍 Auto Server Discovery**: Automatically find DrizLink servers on local network via UDP broadcast
- **💬 Real-time Chat**: Send and receive messages with all connected users
- **🏠 Private Rooms**: Create private chat rooms with selected users for focused collaboration
- **📁 File Sharing**: Transfer files directly between users
- **📂 Folder Sharing**: Share entire folders with other users
- **🔍 File Discovery**: Look up and browse other users' shared directories
- **🔄 Automatic Reconnection**: Seamlessly reconnect with your existing session
- **👥 Status Tracking**: Monitor which users are currently online
- **🎨 Colorful UI**: Enhanced CLI interface with colors and emojis
- **📊 Progress Bars**: Visual feedback for file and folder transfers
- **⏸️ Transfer Controls**: Pause and resume file/folder transfers with unique transfer IDs
- **🔒 Data Integrity**: MD5 checksum verification for files and folders

## 🚀 Installation

### Prerequisites
- Go (1.16 or later) 🔧

### Steps
1. Clone the repository ⬇️
```bash
git clone https://github.com/Harsh2563/DrizLink_Cli.git
cd DrizLink_Cli
```

2. Build the application 🛠️
```bash
go build -o DrizLink_Cli
```

## 🎮 Usage

### Starting the Server 🖥️
```bash
# Start server on default port 8080
go run ./server/cmd --port 8080

# Start server on custom port
go run ./server/cmd --port 3000

```

### Connecting as a Client 📱
```bash
# Auto-discover servers on local network (recommended)
go run ./client/cmd

# Connect to specific server (optional)
go run ./client/cmd --server localhost:8080
go run ./client/cmd --server 192.168.1.100:8080

```

### 🔍 Server Discovery

DrizLink now features **automatic server discovery** via UDP broadcast:
- **No IP sharing required**: Clients automatically scan the local network for available servers
- **Easy selection**: Choose from a list of discovered servers or enter manually
- **Fallback support**: Manual server entry is still available if auto-discovery fails
- **Network scanning**: Servers broadcast their presence every 5 seconds on UDP port 9876

The application will validate:
- Server availability before client connection attempts
- Port availability before starting a server
- Existence of shared folder paths

### 🏠 Room System

DrizLink includes a comprehensive room system for private group communication:
- **Room Creation**: Select from online users to create private rooms
- **Room Management**: Join, leave, and list your rooms easily
- **Context-Aware Chat**: Messages automatically route to your current room
- **Member Control**: Only room members can participate in room conversations
- **Visual Indicators**: Clear UI showing current room status and member counts

## 🏗️ Architecture

The application follows a hybrid P2P architecture:
- 🌐 A central server handles user registration, discovery, and connection brokering
- 📡 UDP broadcast enables automatic server discovery on local networks
- 🏠 Room-based communication allows private group interactions
- ↔️ File and folder transfers occur directly between peers
- 💓 Server maintains connection status through regular heartbeat checks

## 📝 Commands

### Chat Commands 💬
| Command | Description |
|---------|-------------|
| `/help` | Show all available commands |
| `/status` | Show online users |
| `exit` | Disconnect and exit the application |

### Room Commands 🏠
| Command | Description |
|---------|-------------|
| `/createroom` | Create a new room with selected users |
| `/joinroom <roomID>` | Join a specific room |
| `/leaveroom` | Leave current room |
| `/rooms` | List your rooms |

### File Operations 📂
| Command | Description |
|---------|-------------|
| `/lookup <userId>` | Browse user's shared files |
| `/sendfile <userId> <filePath>` | Send a file to another user |
| `/sendfolder <userId> <folderPath>` | Send a folder to another user |
| `/download <userId> <filename>` | Download a file from another user |

### Transfer Controls 📡
| Command | Description |
|---------|-------------|
| `/transfers` | Show all active transfers |
| `/pause <transferId>` | Pause an active transfer |
| `/resume <transferId>` | Resume a paused transfer |

## Terminal UI Features 🎨

- 🌈 **Color-coded messages**:
  - Commands appear in blue
  - Success messages appear in green
  - Error messages appear in red
  - User status notifications in yellow
   - Room messages with special formatting
  
- 📊 **Progress bars for file transfers**:
  ```
  📤 Sending file [===================================>------] 75% (1.2 MB/1.7 MB)
  ```

- 🏠 **Room indicators**:
  ```
  [MyRoom] >>> Hello everyone in this room!
  ```

- 📁 **Improved file listings**:
  ```
  === FOLDERS ===
  📁 [FOLDER] documents (Size: 0 bytes)
  📁 [FOLDER] images (Size: 0 bytes)
  
  === FILES ===
  📄 [FILE] document.pdf (Size: 1024 bytes)
  📄 [FILE] image.jpg (Size: 2048 bytes)
  ```

- 📡 **Transfer management**:
  ```
  📡 Active Transfers:
  ▶ 📤 ID: 1 document.pdf (Active)
     Type: File | Size: 1.2 MB | Progress: 75.0% (900KB/1.2MB)
     To: user123 | Started: 30s ago
  ```

## 🎮 Usage Examples

### Creating and Using Rooms
```bash
# Create a room
/createroom
# Select users: 1,3,5
# Enter room name: "Project Team"

# Join the room
/joinroom room_12345

# Chat in room (messages auto-route to current room)
Hello team! Let's discuss the project.

# Leave room
/leaveroom
```

### File Sharing Workflow
```bash
# Discover what files a user has
/lookup user123

# Send a file
/sendfile user123 /path/to/document.pdf

# Monitor transfer progress
/transfers

# Pause if needed
/pause 1

# Resume transfer
/resume 1
```

## 🔒 Security

The application implements basic reconnection security by tracking IP addresses and user sessions.

- **🔍 Network Discovery**: UDP broadcast messages are limited to local network scope for security
- **🏠 Room Privacy**: Only invited users can join rooms and access room conversations
- **📁 Folder Path Validation**: The application verifies that shared folder paths exist before establishing a connection. If an invalid path is provided, the user will be prompted to enter a valid folder path.
- **🔌 Server Availability Check**: Client automatically verifies server availability before attempting connection, preventing connection errors.
- **🚫 Port Conflict Prevention**: Server detects if a port is already in use and alerts the user to choose another port.
- **📡 Transfer Integrity**: All transfers include unique IDs for tracking and control
- **🔐 Checksum Verification**: All file and folder transfers include MD5 checksum calculation to verify data integrity:
  - When sending, a unique MD5 hash is calculated for the file/folder contents
  - During transfer, the hash is securely transmitted alongside the data
  - Upon receiving, a new hash is calculated from the received data
  - The application compares both hashes to confirm the transfer was successful and uncorrupted
  - Users receive visual confirmation of integrity checks with clear success/failure messages

This checksum process ensures that files and folders arrive exactly as they were sent, protecting against data corruption during transfer.

Made with ❤️ by the DrizLink Team
