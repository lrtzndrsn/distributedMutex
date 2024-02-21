# Distributed Mutual Exclusion

An attempt to solve Distributed Mutual Exclusion as a part of the course Distributed Systems 2023 at ITU. <br>
This is attempted solved by: <br>
Lauritz <lana@itu.dk> <br>
Jonas <kram@itu.dk> <br>
Johan <jsbe@itu.dk> <br>

### Running the program

Standing in the root folder of the project:

Use the following command in the terminal to run the client:
go run Client/client.go -clientArraySize \<client_array_size>

Start multiple instances to simulate a network of clients (our implementation will only work locally on one machine, but can be modified to work globally). To request access to the resource (simulated by a log of the input), enter some text into the terminal en press enter.

## The assignment

### Description:

You have to implement distributed mutual exclusion between nodes in your distributed system. 
Your system has to consist of a set of peer nodes, and you are not allowed to base your implementation on a central server solution.
You can decide to base your implementation on one of the algorithms, that were discussed in lecture 7.

### System Requirements:

- R1: Implement a system with a set of peer nodes, and a Critical Section, that represents a sensitive system operation. Any node can at any time decide it wants access to the Critical Section. Critical section in this exercise is emulated, for example by a print statement, or writing to a shared file on the network.
- R2: Safety: Only one node at the same time is allowed to enter the Critical Section
- R3: Liveliness: Every node that requests access to the Critical Section, will get access to the Critical Section (at some point in time)

### Technical Requirements:

1. Use Golang to implement the service's nodes
2. In you source code repo, provide a README.md, that explains how to start your system
3. Use gRPC for message passing between nodes
4. Your nodes need to find each other. This is called service discovery. You could consider  one of the following options for implementing service discovery:

- Supply a file with IP addresses/ports of other nodes
- Enter IP address/ports through the command line
- use an existing package or service

5. Demonstrate that the system can be started with at least 3 nodesDemonstrate using your system's logs,  a sequence of messages in the system, that leads to a node getting access to the Critical Section. You should provide a discussion of your algorithm, using examples from your logs.

### Hand-in requirements:

1. Hand in a single report in a pdf file. A good report length is between 2-4 pages.

- Describe why your system meets the System Requirements (R1, R2 and R3)
- Provide a discussion of your algorithm, using examples from your logs (Technical Requirement 6)

2. Provide a link to a Git repo with your source code in the report
3. Include system logs, that document the requirements are met, in the appendix of your report
