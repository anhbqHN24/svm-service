# SVM Whiteboard: New User Guidebook

## Introduction
Welcome to the **SVM Whiteboard**! This environment is built to help developers dive deep into the architecture of the Solana Virtual Machine (SVM). Instead of relying strictly on command-line tools, this visual sandbox allows you to write, run, and debug Solana programs directly in your browser. It provides a unified workspace to simulate execution and watch exactly how memory is allocated and accounts are manipulated in real-time, making it an ideal tool for understanding SVM mechanics and testing smart contract logic.

## Platform Overview
The workspace is divided into several distinct modules to give you full visibility into the virtual machine's operations:

1. **Program Editor / Runtime Explorer:** The interface where you write, inspect, and manage your smart contract logic.
2. **Execution Engine:** The control center used to trigger your programs with specific parameters.
3. **Global Memory State & RAM Visualization:** The monitoring dashboard that illustrates blockchain state changes and memory map consumption.
4. **Console Output:** The logging terminal for real-time execution feedback.

---

## How to Use the System Properly

### Step 1: Write or Load Your Program
* Navigate to the **Program Editor**. 
* Input your Solana program logic. Ensure your code is syntactically correct and ready for compilation and simulation within the environment.

### Step 2: Create Accounts
* Because Solana programs are "stateless," they require external Accounts to read from and write to. 
* Before running a program, you need to initialize the necessary accounts. 
* Once created, observe the **SVM Memory (RAM) Visualization** panel. It will transition from *"Memory is empty"* to displaying visual blocks that represent your newly created accounts allocated in the RAM.
* Verify your setup by checking the **Global Memory State** panel to ensure the accounts register as "Loaded".

### Step 3: Configure the Execution Engine
* Go to the **Execution Engine** module.
* **Program Address:** Enter the address of the program you want to test.
* **Parameters:** Feed the required arguments into your program. The engine allows you to set `Param 1` and `Param 2`. You can specify the data type for these parameters, choosing between a `PRIMITIVE` (such as an integer or string) or an `ADDRESS` (a Solana public key).

### Step 4: Execute and Monitor
* Click the **RUN PROGRAM** button in the Execution Engine.
* Immediately check the **Console Output**. The terminal will update from *"Ready for execution..."* to displaying step-by-step logs, success messages, or error tracebacks.
* Watch the **RAM Visualization** and **Global Memory State**. If your program modifies account data or allocates new space, these visualizers will update in real-time to reflect the new state of the SVM.

---

## Best Practices for New Users

* **Start Small:** Write a simple program (like a basic counter or a "Hello World" equivalent) to get used to how the Execution Engine accepts inputs and processes logic.
* **Watch the Memory:** The RAM visualizer is one of the most powerful tools in the whiteboard. Use it to build a strong mental model of *Rent* and *Space* concepts in Solana by seeing exactly how much memory your accounts consume during execution.
* **Debug via Console:** If your program fails to execute, rely heavily on the Console Output. Look out for common Solana errors such as insufficient lamports, incorrect account ownership, or missing signers to rapidly iterate on your code.