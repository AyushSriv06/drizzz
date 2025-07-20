package connection

import (
	"bufio"
	"drizlink/utils"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// Client state
var (
	currentRoomID   string
	currentRoomName string
)
func Connect(address string) (net.Conn, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func Close(conn net.Conn) {
	conn.Close()
}

func UserInput(attribute string, conn net.Conn) error {
	// First check if we get a reconnection signal
	buffer := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, err := conn.Read(buffer)
	conn.SetReadDeadline(time.Time{}) // Reset read deadline

	if err == nil && n > 0 {
		message := string(buffer[:n])
		if strings.HasPrefix(message, "/RECONNECT") {
			parts := strings.SplitN(message, " ", 4)
			if len(parts) == 3 {
				fmt.Printf("Welcome back %s!\n", parts[1])
				return errors.New("reconnect")
			}
		}
	}

	// If no reconnection signal, proceed with normal user input
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Enter your " + attribute + ": ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	// If it's a store file path, validate it
	if attribute == "Store File Path" {
		for {
			// Check if path exists
			if _, err := os.Stat(input); os.IsNotExist(err) {
				fmt.Println(utils.ErrorColor("❌ Error: Directory does not exist")) 
				fmt.Println("Enter a valid " + attribute + ": ")
				input, _ = reader.ReadString('\n')
				input = strings.TrimSpace(input)
				continue
			}

			// Check if it's a directory
			fileInfo, err := os.Stat(input)
			if err != nil || !fileInfo.IsDir() {
				fmt.Println(utils.ErrorColor("❌ Error: Path is not a directory"))
				fmt.Println("Enter a valid " + attribute + ": ")
				input, _ = reader.ReadString('\n')
				input = strings.TrimSpace(input)
				continue
			}

			break
		}
	}

	_, err = conn.Write([]byte(input))
	if err != nil {
		fmt.Println("error in write " + attribute)
		panic(err)
	}

	return nil
}

func ReadLoop(conn net.Conn) {
	for {
		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Println(utils.ErrorColor("❌ Connection lost:"), err)
			return
		}
		message := string(buffer[:n])
		switch {
		case strings.HasPrefix(message, "/FILE_RESPONSE"):
			fmt.Println(utils.InfoColor("📥 File transfer starting..."))
			args := strings.SplitN(message, " ", 5)
			if len(args) != 5 {
				fmt.Println(utils.ErrorColor("❌ Invalid arguments. Use: /FILE_RESPONSE <userId> <filename> <fileSize> <storeFilePath>"))
				continue
			}
			recipientId := args[1]
			fileName := args[2]
			fileSizeStr := strings.TrimSpace(args[3])
			fileSize, err := strconv.ParseInt(fileSizeStr, 10, 64)
			storeFilePath := args[4]
			if err != nil {
				fmt.Println(utils.ErrorColor("❌ Invalid fileSize. Use: /FILE_RESPONSE <userId> <filename> <fileSize> <storeFilePath>"))
				continue
			}

			HandleFileTransfer(conn, recipientId, fileName, int64(fileSize), storeFilePath)
			continue
		case strings.HasPrefix(message, "/FOLDER_RESPONSE"):
			fmt.Println(utils.InfoColor("📥 Folder transfer starting..."))
			args := strings.SplitN(message, " ", 5)
			if len(args) != 5 {
				fmt.Println(utils.ErrorColor("❌ Invalid arguments. Use: /FOLDER_RESPONSE <userId> <folderName> <folderSize> <storeFilePath>"))
				continue
			}
			recipientId := args[1]
			folderName := args[2]
			folderSizeStr := strings.TrimSpace(args[3])
			folderSize, err := strconv.ParseInt(folderSizeStr, 10, 64)
			storeFilePath := args[4]
			if err != nil {
				fmt.Println(utils.ErrorColor("❌ Invalid folderSize. Use: /FOLDER_RESPONSE <userId> <folderName> <folderSize> <storeFilePath>"))
				continue
			}
			HandleFolderTransfer(conn, recipientId, folderName, folderSize, storeFilePath)
			continue
		case strings.HasPrefix(message, "ONLINE_USERS_LIST"):
			// Handle online users list for room creation
			handleOnlineUsersList(message)
			continue
		case strings.HasPrefix(message, "ROOM_CREATED"):
			args := strings.SplitN(message, " ", 4)
			if len(args) >= 4 {
				roomID := args[1]
				roomName := args[2]
				creatorName := args[3]
				fmt.Printf("%s Room '%s' created by %s (ID: %s)\n", 
					utils.SuccessColor("🏠"), 
					utils.InfoColor(roomName), 
					utils.UserColor(creatorName),
					utils.CommandColor(roomID))
				fmt.Printf("  Use %s to join this room\n", utils.CommandColor("/joinroom "+roomID))
			}
			continue
		case strings.HasPrefix(message, "ROOM_JOINED"):
			args := strings.SplitN(message, " ", 3)
			if len(args) >= 3 {
				roomID := args[1]
				roomName := args[2]
				currentRoomID = roomID
				currentRoomName = roomName
				fmt.Printf("%s Joined room '%s' (ID: %s)\n", 
					utils.SuccessColor("✅"), 
					utils.InfoColor(roomName),
					utils.CommandColor(roomID))
				fmt.Printf("  Messages will now be sent to this room. Use %s to leave.\n", 
					utils.CommandColor("/leaveroom"))
			}
			continue
		case strings.HasPrefix(message, "ROOM_LEFT"):
			args := strings.SplitN(message, " ", 2)
			if len(args) >= 2 {
				roomID := args[1]
				fmt.Printf("%s Left room %s\n", 
					utils.WarningColor("👋"), 
					utils.CommandColor(roomID))
				currentRoomID = ""
				currentRoomName = ""
				fmt.Println(utils.InfoColor("  Back to general chat"))
			}
			continue
		case strings.HasPrefix(message, "ROOMS_LIST"):
			handleRoomsList(message)
			continue
		case message == "ROOM_NOT_FOUND":
			fmt.Println(utils.ErrorColor("❌ Room not found"))
			continue
		case message == "NOT_ROOM_MEMBER":
			fmt.Println(utils.ErrorColor("❌ You are not a member of this room"))
			continue
		case strings.HasPrefix(message, "PING"):
			_, err = conn.Write([]byte("PONG\n"))
			if err != nil {
				fmt.Println(utils.ErrorColor("❌ Error responding to heartbeat:"), err)
				continue
			}
		case message == "USERS:":
			// Improved approach to accumulate the complete user list
			fmt.Println(utils.HeaderColor("\n👥 Online Users:"))
			fmt.Println(utils.InfoColor("-------------------"))

			// Read the complete user list with timeout
			userList := ""
			tempBuf := make([]byte, 1024)
			conn.SetReadDeadline(time.Now().Add(2 * time.Second))

			for {
				m, err := conn.Read(tempBuf)
				if err != nil {
					break // Break on error (likely timeout)
				}
				userList += string(tempBuf[:m])
				if m < 1024 {
					break // All data received
				}
			}

			// Reset the deadline
			conn.SetReadDeadline(time.Time{})

			// Process users
			userCount := 0
			for _, line := range strings.Split(userList, "\n") {
				if strings.TrimSpace(line) != "" {
					userCount++
					// Enhanced formatting for username and ID
					if strings.Contains(line, "[ID:") {
						parts := strings.SplitN(line, "[ID:", 2)
						if len(parts) == 2 {
							username := strings.TrimSpace(parts[0])
							idPart := strings.SplitN(parts[1], "]", 2)
							if len(idPart) == 2 {
								userId := strings.TrimSpace(idPart[0])
								status := strings.TrimSpace(idPart[1])
								fmt.Printf("%s %s %s %s %s\n",
									utils.SuccessColor(" •"),
									utils.UserColor(username),
									utils.InfoColor("(ID:"),
									utils.CommandColor(userId),
									utils.InfoColor(")"+status))
								continue
							}
						}
					}
					// Fallback to original formatting if parsing fails
					fmt.Println(utils.SuccessColor(" • "), utils.UserColor(line))
				}
			}

			if userCount == 0 {
				fmt.Println(utils.InfoColor(" No users currently online"))
			}

			fmt.Println(utils.InfoColor("-------------------"))
			continue
		case strings.HasPrefix(message, "/LOOK_REQUEST"):
			args := strings.SplitN(message, " ", 3)
			if len(args) != 3 {
				fmt.Println(utils.ErrorColor("❌ Invalid arguments. Use: /LOOK_REQUEST <storageFilePath> <userId>"))
				continue
			}
			storageFilePath := args[2]
			userId := args[1]
			fmt.Println(utils.InfoColor("🔍 Processing directory lookup request from"), utils.UserColor(userId))
			HandleLookupResponse(conn, storageFilePath, userId)
			continue
		case strings.HasPrefix(message, "/LOOK_RESPONSE"):
			args := strings.SplitN(message, " ", 3)
			if len(args) != 3 {
				fmt.Println(utils.ErrorColor("❌ Invalid arguments. Use: /LOOK_RESPONSE <userId> <files>"))
				continue
			}
			userId := args[1]
			files := strings.Split(args[2], " ")

			fmt.Println(utils.HeaderColor("\n📂 Directory Listing for User:"), utils.UserColor(userId))
			fmt.Println(utils.InfoColor("-------------------------------------------"))

			for _, file := range files {
				if strings.HasPrefix(file, "[FOLDER]") {
					fmt.Println(utils.WarningColor("📁"), utils.InfoColor(file))
				} else if strings.HasPrefix(file, "[FILE]") {
					fmt.Println(utils.SuccessColor("📄"), utils.InfoColor(file))
				} else if strings.HasPrefix(file, "===") {
					fmt.Println(utils.HeaderColor(file))
				} else {
					fmt.Println(utils.InfoColor(file))
				}
			}

			fmt.Println(utils.InfoColor("-------------------------------------------\n"))
			continue
		case strings.HasPrefix(message, "/DOWNLOAD_REQUEST"):
			args := strings.SplitN(message, " ", 3)
			if len(args) != 3 {
				fmt.Println(utils.ErrorColor("❌ Invalid arguments. Use: /DOWNLOAD_REQUEST <userId> <filename>"))
				continue
			}
			userId := args[1]
			filePath := args[2]
			fmt.Println(utils.InfoColor("📤 Download request from"), utils.UserColor(userId), utils.InfoColor("for"), utils.InfoColor(filePath))
			HandleDownloadResponse(conn, userId, filePath)
			continue
		default:
			if strings.HasPrefix(message, "[Room ") {
				// Room message
				fmt.Println(utils.InfoColor(message))
			} else if strings.Contains(message, "has joined the chat") {
				fmt.Println(utils.WarningColor("👋 " + message))
			} else if strings.Contains(message, "has rejoined the chat") {
				fmt.Println(utils.WarningColor("🔄 " + message))
			} else if strings.Contains(message, "is now offline") {
				fmt.Println(utils.WarningColor("👋 " + message))
			} else {
				fmt.Println(message)
			}
		}
	}
}

func handleOnlineUsersList(message string) {
	parts := strings.SplitN(message, " ", 2)
	if len(parts) < 2 {
		fmt.Println(utils.ErrorColor("❌ No users available for room creation"))
		return
	}
	
	userPairs := strings.Split(parts[1], " ")
	if len(userPairs) == 0 {
		fmt.Println(utils.ErrorColor("❌ No other users online"))
		return
	}
	
	fmt.Println(utils.HeaderColor("\n👥 Select users for the room:"))
	fmt.Println(utils.InfoColor("--------------------------------"))
	
	var users []struct {
		ID   string
		Name string
	}
	
	for _, pair := range userPairs {
		if pair == "" {
			continue
		}
		parts := strings.Split(pair, "|")
		if len(parts) == 2 {
			users = append(users, struct {
				ID   string
				Name string
			}{parts[0], parts[1]})
			fmt.Printf("%s %s %s %s\n", 
				utils.CommandColor(fmt.Sprintf("[%d]", len(users))),
				utils.UserColor(parts[1]),
				utils.InfoColor("(ID:"),
				utils.InfoColor(parts[0]+")"))
		}
	}
	
	if len(users) == 0 {
		fmt.Println(utils.ErrorColor("❌ No other users online"))
		return
	}
	
	fmt.Println(utils.InfoColor("--------------------------------"))
	fmt.Print(utils.CommandColor("Enter room name: "))
	
	reader := bufio.NewReader(os.Stdin)
	roomName, _ := reader.ReadString('\n')
	roomName = strings.TrimSpace(roomName)
	
	if roomName == "" {
		fmt.Println(utils.ErrorColor("❌ Room name cannot be empty"))
		return
	}
	
	fmt.Print(utils.CommandColor("Select users (comma-separated numbers, e.g., 1,3,5): "))
	selection, _ := reader.ReadString('\n')
	selection = strings.TrimSpace(selection)
	
	if selection == "" {
		fmt.Println(utils.ErrorColor("❌ No users selected"))
		return
	}
	
	selectedNumbers := strings.Split(selection, ",")
	var selectedUserIDs []string
	
	for _, numStr := range selectedNumbers {
		numStr = strings.TrimSpace(numStr)
		num, err := strconv.Atoi(numStr)
		if err != nil || num < 1 || num > len(users) {
			fmt.Printf("%s Invalid selection: %s\n", utils.ErrorColor("❌"), numStr)
			continue
		}
		selectedUserIDs = append(selectedUserIDs, users[num-1].ID)
	}
	
	if len(selectedUserIDs) == 0 {
		fmt.Println(utils.ErrorColor("❌ No valid users selected"))
		return
	}
	
	// Send room creation request
	createRoomMsg := fmt.Sprintf("/CREATE_ROOM %s %s", roomName, strings.Join(selectedUserIDs, ","))
	// This will be sent via the connection in WriteLoop
	fmt.Printf("%s Creating room '%s' with %d users...\n", 
		utils.InfoColor("🏠"), 
		utils.InfoColor(roomName), 
		len(selectedUserIDs))
	
	// We need to send this message through the connection
	// This is a bit tricky since we're in ReadLoop, but we'll handle it in WriteLoop
	pendingRoomCreation = createRoomMsg
}

func handleRoomsList(message string) {
	parts := strings.SplitN(message, " ", 2)
	if len(parts) < 2 || strings.TrimSpace(parts[1]) == "" {
		fmt.Println(utils.InfoColor("📭 You are not a member of any rooms"))
		return
	}
	
	roomPairs := strings.Split(parts[1], " ")
	
	fmt.Println(utils.HeaderColor("\n🏠 Your Rooms:"))
	fmt.Println(utils.InfoColor("---------------"))
	
	for _, pair := range roomPairs {
		if pair == "" {
			continue
		}
		parts := strings.Split(pair, "|")
		if len(parts) == 3 {
			roomID := parts[0]
			roomName := parts[1]
			memberCount := parts[2]
			
			status := ""
			if roomID == currentRoomID {
				status = utils.SuccessColor(" [CURRENT]")
			}
			
			fmt.Printf("%s %s %s %s%s\n", 
				utils.InfoColor("🏠"),
				utils.InfoColor(roomName),
				utils.CommandColor("(ID: "+roomID+")"),
				utils.InfoColor("Members: "+memberCount),
				status)
		}
	}
	fmt.Println(utils.InfoColor("---------------"))
	fmt.Printf("Use %s to join a room\n", utils.CommandColor("/joinroom <roomID>"))
}

var pendingRoomCreation string
func WriteLoop(conn net.Conn) {
	reader := bufio.NewReader(os.Stdin)
	for {
		// Check for pending room creation
		if pendingRoomCreation != "" {
			_, err := conn.Write([]byte(pendingRoomCreation))
			if err != nil {
				fmt.Println(utils.ErrorColor("❌ Error creating room:"), err)
			}
			pendingRoomCreation = ""
			continue
		}
		
		// Show current context in prompt
		prompt := ">>> "
		if currentRoomID != "" {
			prompt = fmt.Sprintf("[%s] >>> ", utils.InfoColor(currentRoomName))
		}
		fmt.Print(utils.CommandColor(prompt))
		
		fmt.Print(utils.CommandColor(">>> "))
		message, _ := reader.ReadString('\n')
		message = strings.TrimSpace(message)
		switch {
		case message == "exit":
			fmt.Println(utils.InfoColor("👋 Goodbye!"))
			conn.Close()
			return
		case message == "/help":
			utils.PrintHelp()
			continue
		case message == "/createroom":
			fmt.Println(utils.InfoColor("🏠 Fetching online users..."))
			_, err := conn.Write([]byte("/GET_ONLINE_USERS"))
			if err != nil {
				fmt.Println(utils.ErrorColor("❌ Error fetching users:"), err)
			}
			continue
		case strings.HasPrefix(message, "/joinroom"):
			args := strings.SplitN(message, " ", 2)
			if len(args) != 2 {
				fmt.Println(utils.ErrorColor("❌ Invalid arguments. Use: /joinroom <roomID>"))
				continue
			}
			roomID := strings.TrimSpace(args[1])
			_, err := conn.Write([]byte(fmt.Sprintf("/JOIN_ROOM %s", roomID)))
			if err != nil {
				fmt.Println(utils.ErrorColor("❌ Error joining room:"), err)
			}
			continue
		case message == "/leaveroom":
			_, err := conn.Write([]byte("/LEAVE_ROOM"))
			if err != nil {
				fmt.Println(utils.ErrorColor("❌ Error leaving room:"), err)
			}
			continue
		case message == "/rooms":
			_, err := conn.Write([]byte("/LIST_ROOMS"))
			if err != nil {
				fmt.Println(utils.ErrorColor("❌ Error fetching rooms:"), err)
			}
			continue
		case strings.HasPrefix(message, "/sendfile"):
			args := strings.SplitN(message, " ", 3)
			if len(args) != 3 {
				fmt.Println(utils.ErrorColor("❌ Invalid arguments. Use: /sendfile <userId> <filename>"))
				continue
			}
			recipientId := args[1]
			filePath := args[2]
			fmt.Println(utils.InfoColor("📤 Sending file to"), utils.UserColor(recipientId))
			HandleSendFile(conn, recipientId, filePath)
			continue
		case strings.HasPrefix(message, "/sendfolder"):
			args := strings.SplitN(message, " ", 3)
			if len(args) != 3 {
				fmt.Println(utils.ErrorColor("❌ Invalid arguments. Use: /sendfolder <userId> <folderPath>"))
				continue
			}
			recipientId := args[1]
			folderPath := args[2]
			fmt.Println(utils.InfoColor("📤 Sending folder to"), utils.UserColor(recipientId))
			HandleSendFolder(conn, recipientId, folderPath)
			continue
		case strings.HasPrefix(message, "/lookup"):
			args := strings.SplitN(message, " ", 2)
			if len(args) != 2 {
				fmt.Println(utils.ErrorColor("❌ Invalid arguments. Use: /lookup <userId>"))
				continue
			}
			recipientId := args[1]
			fmt.Println(utils.InfoColor("🔍 Looking up files for user"), utils.UserColor(recipientId))
			HandleLookupRequest(conn, recipientId)
			continue
		case strings.HasPrefix(message, "/status"):
			fmt.Println(utils.InfoColor("👥 Fetching online users..."))
			_, err := conn.Write([]byte(message))
			if err != nil {
				fmt.Println(utils.ErrorColor("❌ Error checking status:"), err)
				continue
			}
			continue
		case strings.HasPrefix(message, "/download"):
			args := strings.SplitN(message, " ", 3)
			if len(args) != 3 {
				fmt.Println(utils.ErrorColor("❌ Invalid arguments. Use: /download <userId> <filename>"))
				continue
			}
			recipientId := args[1]
			filePath := args[2]
			fmt.Println(utils.InfoColor("📥 Requesting download from"), utils.UserColor(recipientId))
			HandleDownloadRequest(conn, recipientId, filePath)
			continue
		case strings.HasPrefix(message, "/transfers"):
			HandleListTransfers()
			continue
		case strings.HasPrefix(message, "/pause"):
			args := strings.SplitN(message, " ", 2)
			if len(args) != 2 {
				fmt.Println(utils.ErrorColor("❌ Invalid arguments. Use: /pause <transferId>"))
				continue
			}
			transferID := args[1]
			HandlePauseTransfer(transferID)
			continue
		case strings.HasPrefix(message, "/resume"):
			args := strings.SplitN(message, " ", 2)
			if len(args) != 2 {
				fmt.Println(utils.ErrorColor("❌ Invalid arguments. Use: /resume <transferId>"))
				continue
			}
			transferID := args[1]
			HandleResumeTransfer(transferID)
			continue
		default:
			if message != "" {
				// If in a room, send as room message
				if currentRoomID != "" {
					message = fmt.Sprintf("/ROOM_MESSAGE %s %s", currentRoomID, message)
				}
				_, err := conn.Write([]byte(message))
				if err != nil {
					fmt.Println(utils.ErrorColor("❌ Error sending message:"), err)
					return
				}
			}
		}
	}
}
