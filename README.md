# Solana Modular SVM Whiteboard

## Introduction

**SVM Whiteboard** is a simple educational project to learn how **Solana Virtual Machine (SVM)** works. It is written in **Go** and simulates the core logic of Solana, like accounts, instructions, and memory management.

> **Note:** This is for learning purposes, not a real blockchain node.

🔗 **Slides:** [Project Introduction Slideshow](https://docs.google.com/presentation/d/1ltbNwZAnafnMy9UtMbuVjRLkNS0z7ARLzaD6FXHyi8Y/view?usp=sharing)

## 🏗 Architecture
<img width="560" height="591" alt="Screenshot 2026-01-27 at 10 00 13 AM" src="https://github.com/user-attachments/assets/26f8e709-3e6c-4f3e-909e-d3004eceb4db" />


The project follows a simple structure:

- **`cmd/`**: Main entry point to start the server.
- **`app/api/`**: Handles HTTP requests (read, write, execute).
- **`app/compiler/`**: Compiles high-level source code into executable bytecode.
- **`runtime/`**: Manages the Memory and Accounts (VM In-memory like simulation) with session support.
- **`app/program/`**: The Virtual Machine (VM) that runs bytecode.
- **`app/model/`**: Basic model definitions (Opcodes, Account types, etc,...).
- **`app/service/`**: Business logic layer for account operations and program execution.
- **`helper/`**: Utility functions for serialization, encoding, and common operations.

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

## 🔐 Session Management

The SVM Whiteboard supports **multi-user sessions** where each user has their own isolated environment:

- **Session ID**: Each client can send a unique `X-Session-ID` header to maintain their own VM state
- **Isolated Memory**: Each session has its own Genesis account, account storage, and program cache
- **Default Session**: If no session ID is provided, requests use a shared "default_public" session
- **Concurrent Access**: Multiple users can interact with the system simultaneously without interference

### Example with Session ID:
```bash
curl -H "X-Session-ID: user123" http://localhost:9924/all
```

## 💻 Compiler & High-Level Language

The SVM Whiteboard includes a **compiler** that translates human-readable source code into executable bytecode. This makes it much easier to write programs compared to manually crafting bytecode arrays.

### Compile Source Code to Bytecode:
```bash
curl -X POST http://localhost:9924/compile \
 -H "Content-Type: application/json" \
 -d '{
       "source_code": "SET x 10\nSET y 20\nADD x y\nPRINT_INT x\nHALT"
     }'
```

The compiler will return a binary file (`program.bin`) that can be deployed as an executable account.

### Supported High-Level Instructions:

#### **Data Movement**
- `SET dest value` - Assign a value to a variable (supports numbers, strings, and variables)
  - Example: `SET x 10`, `SET msg "hello"`, `SET y x`

#### **Arithmetic Operations** 
- `ADD dest src` - Add src to dest (dest = dest + src)
- `SUB dest src` - Subtract src from dest (dest = dest - src)
- `MUL dest src` - Multiply dest by src (dest = dest * src)
- `DIV dest src` - Divide dest by src (dest = dest / src)
- `MOD dest src` - Modulo operation (dest = dest % src)

#### **Logical Operations**
- `AND dest src` - Bitwise AND operation
- `OR dest src` - Bitwise OR operation
- `XOR dest src` - Bitwise XOR operation

#### **String Operations**
- `CONCAT dest src` - Concatenate strings (dest = dest + src)

#### **Comparison & Control Flow**
- `CMP op1 op2` - Compare two values and set Flag register (-1: less, 0: equal, 1: greater)
- `JMP addr` - Unconditional jump to instruction address
- `JEQ addr` - Jump if equal (Flag == 0)
- `JNE addr` - Jump if not equal (Flag != 0)
- `JGT addr` - Jump if greater (Flag == 1)
- `JLT addr` - Jump if less (Flag == -1)

#### **Input/Output**
- `PRINT_INT var` - Print integer value
- `PRINT_STR var` - Print string value
- `HALT` - Stop program execution

### Example Programs:

**Simple Addition:**
```
SET x 10
SET y 20
ADD x y
PRINT_INT x
HALT
```

**String Manipulation:**
```
SET msg "Hello"
SET world " World"
CONCAT msg world
PRINT_STR msg
HALT
```

**Conditional Logic:**
```
SET a 15
SET b 10
CMP a b
JGT 9
PRINT_INT b
HALT
PRINT_INT a
HALT
```

### Reserved Variables:
- `param_1` - Always mapped to Stack[0] (first execution parameter)
- `param_2` - Always mapped to Stack[1] (second execution parameter)
- User variables start from Stack[2] onwards

### Compilation Process:
1. **Lexical Analysis**: Parse source code into tokens
2. **Symbol Table**: Track variable names and their stack addresses
3. **Code Generation**: Convert high-level instructions to bytecode
4. **Optimization**: Automatically append HALT instruction if missing

## 📖 How to Use (REST API)
1. **Read an Account (GET)**
Get Account State (JSON object containing Account metadata) from Account Address
    ```Bash
    curl http://localhost:9924/read/{GENESIS_ACCOUNT_ADDRESS}
    ```
2. **Create an Account (POST)**
Create & Deploy a new program or data Account.

For **Program Accounts** (executable), you can use:
- **Compiled bytecode**: First compile your source code using `/compile`, then use the hex string
- **Raw bytecode array**: Manually provide bytecode as hex string

**Example with compiled bytecode:**
```bash
# Step 1: Compile source code
curl -X POST http://localhost:9924/compile \
 -H "Content-Type: application/json" \
 -d '{"source_code": "ADD param_1 param_2\nPRINT_INT param_1\nHALT"}' \
 --output program.bin

# Step 2: Convert to hex and deploy
BYTECODE_HEX=$(xxd -p program.bin | tr -d '\n')
curl -X POST http://localhost:9924/write \
 -H "Content-Type: application/json" \
 -d "{
       \"owner\": \"{GENESIS_ACCOUNT_ADDRESS}\",
       \"executable\": true,
       \"data\": \"$BYTECODE_HEX\"
     }"
```

**Example with raw bytecode:**
```bash
curl -X POST http://localhost:9924/write \
 -H "Content-Type: application/json" \
 -d '{
       "owner": "{GENESIS_ACCOUNT_ADDRESS}",
       "executable": true,
       "data": "0102001000010ee000ff0000"
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

## 🧩 Low-Level Opcodes Reference

> **Note:** For everyday use, prefer the [High-Level Language](#-compiler--high-level-language) which compiles to these opcodes automatically.

These are the raw bytecode instructions that the VM executes:

| Opcode | Name | Description |
| :--- | :--- | :--- |
| **Data Movement** | | |
| `0x01` | `LOAD_IMM` | Load immediate value into register |
| `0x02` | `MOV` | Copy value from source register to destination register |
| `0x05` | `LOAD_STR` | Load string from Heap to register |
| `0x06` | `LOAD` | Load from Stack memory to register |
| `0x07` | `STORE` | Store from register to Stack memory |
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
| **String Operations** | | |
| `0x40` | `CONCAT` | Concatenate strings from source to destination register |
| **IO & System** | | |
| `0xEE` | `PRINT_INT` | Print integer value from register |
| `0xF0` | `PRINT_STR` | Print string from memory address (Pointer) in register |
| `0xFF` | `HALT` | Halt execution |

## 🤝 Contributing
This project is for learning. Feel free to open a Pull Request if you want to improve the code!

You can reach out to me via telegram ```t.me/lucasb_hn24```. With love, Lucas Bui.
