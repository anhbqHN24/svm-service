# Solana Modular SVM Whiteboard

## Introduction

**SVM Whiteboard** is a simple educational project to learn how **Solana Virtual Machine (SVM)** works. It is written in **Go** and simulates the core logic of Solana, like accounts, instructions, and memory management.

> **Note:** This is for learning purposes, not a real blockchain node.

🔗 **Slides:** [Project Introduction Slideshow](https://docs.google.com/presentation/d/1ltbNwZAnafnMy9UtMbuVjRLkNS0z7ARLzaD6FXHyi8Y/view?usp=sharing)

## 🏗 Architecture
<img width="1236" height="755" alt="Screenshot 2026-01-20 at 8 24 33 PM" src="https://github.com/user-attachments/assets/88a10191-5e11-4434-8558-efee1f997407" />


The project follows a simple structure:

- **`cmd/`**: Main entry point to start the server.
- **`app/api/`**: Handles HTTP requests (read, write, execute).
- **`runtime/`**: Manages the Memory and Accounts (VM In-memory like simulation).
- **`app/program/`**: The Virtual Machine (VM) that runs bytecode.
- **`app/core/`**: Basic model definitions (Opcodes, Account types, etc,...).

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
Server runs on port :9924

## 📖 How to Use
1. **Read an Account (GET)**
Check the terminal logs for the Genesis Account Structure after starting the server.
    ```Bash
    curl http://localhost:9924/read/{GENESIS_ACCOUNT_ADDRESS}
    ```
2. **Create an Account (POST)**
Create & Deploy a new program or data Account.
    ```Bash
    curl -X POST http://localhost:9924/write \
     -H "Content-Type: application/json" \
     -d '{
           "owner": {PROGRAM_ACCOUNT_ADDRESS},
           "executable": true,
           "data": [1, 2, 0, 16, 0, 1, 238, 0, 0, 255, 0, 0]
         }'
    ```

3. **Execute a Program (POST)**
Run the program you just created. This action will be divided into two cases, depending on the involvement of the "data_addr" variable in the request BODY :
- included: binary execution will take data of Data Account as one of the parameter & param_1 as the other -> update the Account data (```Account state``` in Solana)
- not included: binary execution will take 2 param from "params" variable inside request BODY -> run the code as normal.
    ```Bash
    curl -X POST http://localhost:9924/execute \
     -H "Content-Type: application/json" \
     -d '{
           "prog_addr": "{YOUR_PROGRAM_ACCOUNT_ADDRESS}",
           "data_addr": "{YOUR_DATA_ACCOUNT_ADDRESS}",
           "params": {
             "param_1": 10,
             "param_2": 20
           }
         }'
    ```
Please noted that this SVM only accept 2 parameters to execute program Account at the moment for simplicity.

## 🧩 Supported Opcodes

| Opcode | Name | Description |
| :--- | :--- | :--- |
| **Data Movement** | | |
| `0x01` | `LOAD_IMM` | Load a number into a register |
| `0x02` | `MOV` | Copy value from source register to destination register |
| **Arithmetic** | | |
| `0x10` | `ADD` | Add source to destination |
| `0x11` | `SUB` | Subtract source from destination |
| `0x12` | `MUL` | Multiply source and destination |
| `0x13` | `DIV` | Divide destination by source (Checks zero division) |
| `0x14` | `MOD` | Modulo destination by source (Checks zero division) |
| **Logic** | | |
| `0x20` | `AND` | Bitwise AND operation |
| `0x21` | `OR` | Bitwise OR operation |
| `0x22` | `XOR` | Bitwise XOR operation |
| **Control Flow** | | |
| `0x30` | `CMP` | Compare RegA and RegB, set Flag (-1, 0, 1) |
| `0x31` | `JMP` | Unconditional jump to address |
| `0x32` | `JEQ` | Jump if Flag == 0 (Equal) |
| `0x33` | `JNE` | Jump if Flag != 0 (Not Equal) |
| `0x34` | `JGT` | Jump if Flag == 1 (Greater) |
| `0x35` | `JLT` | Jump if Flag == -1 (Less) |
| **IO & System** | | |
| `0xEE` | `PRINT_INT` | Print integer value from register |
| `0xF0` | `PRINT_STR` | Print string from memory address (Pointer) in register |
| `0xFF` | `HALT` | Halt execution |

## 🤝 Contributing
This project is for learning. Feel free to open a Pull Request if you want to improve the code!

You can reach out to me via telegram ```t.me/lucasb_hn24```. With love, Lucas Bui.

