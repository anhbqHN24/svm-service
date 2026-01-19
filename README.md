# Solana Modular SVM Whiteboard

## Introduction

**SVM Whiteboard** is a simple educational project to learn how **Solana Virtual Machine (SVM)** works. It is written in **Go** and simulates the core logic of Solana, like accounts, instructions, and memory management.

> **Note:** This is for learning purposes, not a real blockchain node.

🔗 **Slides:** [Project Introduction Slideshow](https://docs.google.com/presentation/d/1ltbNwZAnafnMy9UtMbuVjRLkNS0z7ARLzaD6FXHyi8Y/view?usp=sharing)

## 🏗 Architecture

![Architecture Diagram](https://drive.google.com/file/d/13fUoZKP6mkAdkKC2tZ2-eONd8zrZ4Ctk)

The project follows a simple structure:

- **`cmd/`**: Main entry point to start the server.
- **`api/`**: Handles HTTP requests (read, write, execute).
- **`internal/runtime/`**: Manages the Ledger and Accounts (VM In-memory like simulation).
- **`internal/program/`**: The Virtual Machine (VM) that runs bytecode.
- **`internal/core/`**: Basic model definitions (Opcodes, Account types, etc,...).

## 🚀 How to Run

### Prerequisites
- Go 1.24 or higher

### Steps
1. **Clone the repo:**
   ```Bash
   git clone [https://github.com/yourusername/svm_whiteboard.git](https://github.com/yourusername/svm_whiteboard.git)
   cd svm_whiteboard
   ```
2. **Install dependencies:**
    ```Bash
    go mod tidy
    ```
3. **Start the Server:**
    ```Bash
    go run cmd/server.go
    ```
Server runs on port ```:9924```

## 📖 How to Use
1. **Read an Account (GET)**
Get Account State (JSON object containing Account metadata) from Account Address
    ```Bash
    curl http://localhost:9924/read/{GENESIS_ACCOUNT_ADDRESS}
    ```
2. **Create an Account (POST)**
Deploy a new program or data account.
    ```Bash
    curl -X POST http://localhost:9924/write \
     -H "Content-Type: application/json" \
     -d '{
           "owner_id": "BPFLoader_1",
           "executable": true,
           "data": [1, 2, 0, 16, 0, 1, 238, 0, 0, 255, 0, 0]
         }'
    ```

3. **Execute a Program (POST)**
Run the program you just created.
    ```Bash
    curl -X POST http://localhost:9924/execute \
     -H "Content-Type: application/json" \
     -d '{
           "address": "{YOUR_PROGRAM_ACCOUNT_ADDRESS}",
           "params": {
             "param_1": 10,
             "param_2": 20
           }
         }'
    ```
Please noted that this SVM only accept 2 parameters to execute program Account at the moment for simplicity.

## 🧩 Supported Opcodes
| Opcode | Name | Description |
|---|---|---|
| 0x01 | LOAD_IMM | Load a number into a register |

## 🤝 Contributing
This project is for learning. Feel free to open a Pull Request if you want to improve the code!

You can reach out to me via telegram ```t.me/lucasb_hn24```. With love, Lucas Bui.

