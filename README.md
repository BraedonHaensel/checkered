# Checkered

CPSC 559 - Group Project - Winter 2026

Checkered is an online, distributed platform for the classic board game [checkers](https://www.youtube.com/watch?v=MOW9k_C4vFU).

This application is for both casual players who want to play for fun and competitive players aiming to climb the leaderboards!

# Group Members

- Braedon Haensel
- Brian Heckel
- Avery Keuben
- Pranab Mainali
- Taylor Wong

## Project Demo

For a brief introduction to our application and its underlying distributed system, please see our [presentation video](documents/checkered-10-minute-presentation.mp4).

## Process types

The distributed system for Checkered involves the following process types:

1.  Name Server
    - One server running on a static IP
    - Used to dynamically register, retrieve, and deregister servers in the distributed system

2.  Matchmakers (Cluster)
    - Manage the queue of waiting players (clients)
    - Assign matches to Game Servers
    - Store the leaderboard state
    - One Matchmaker is elected as a leader of the cluster using a Bully algorithm

3.  Game Servers (Cluster)
    - Host matches between players
    - Manage the logic for ongoing checkers games

4.  Clients
    - End users connected to the distributed system
    - Client software handles the UI display and communication with the distributed system

## How to Run the Application

### Step 1: Clone the repository

Clone the Checkered repository with this command:

```bash
git clone https://github.com/akeuben/Checkered.git
```

Navigate into the repository:

```bash
cd CPSC-559-Group-Project
```

### Step 2: Create the .env file

Create the `.env` file by copying the `.env.example` file:

```bash
cp .env.example .env
```

_Note_: Change the `APP_NAMESERVER_URL` environment variable in `.env` file to connect to a Name Server running on a different address or machine.

### Step 3: Start the System Processes

#### Option 1: All-In-One Startup

Start all of our services together:

```bash
./start.sh
```

_Tip_: To run shell scripts on Windows, use [Git Bash](https://gitforwindows.org/) or a [WSL](https://learn.microsoft.com/en-us/windows/wsl/install) terminal.

#### Option 2: Manual Startup

Alternatively, each process can be started up manually.

##### Name Server

Run the Name Server with this command:

```bash
cd name-server && go run .
```

_Note_: There should only be one Name Server.

_Tip_: Change the `APP_NAMESERVER_URL` environment variable in the `.env` file to update the Name Server address.

##### Matchmakers

Run each Matchmaker with this command:

```bash
cd backend/cmd/matchmaker && go run . -addr :4000
```

_Tip_: Increment the port for each additional Matchmaker.

##### Game Servers

Run each Game Server with this command:

```bash
cd backend/cmd/game-server && go run . -addr :5000
```

_Tip_: Increment the port for each additional Game Server.

##### Frontend

To run the frontend:

1. Navigate into the `frontend` directory:

    ```bash
    cd frontend
    ```

2. Use [Node.js](https://nodejs.org/en) to install the required npm packages:

    ```bash
    npm install
    ```

3. Build the frontend for a _production_ environment:

    ```bash
    npm run build
    ```

4. Preview the _production_ build locally:

    ```bash
    npm run preview
    ```

    - Alternatively, preview the _production_ build on your local network so clients on different machines can connect:

        ```bash
        npm run preview:host
        ```

_Note_: For a _development_ environment, run `npm run dev` after installing the required npm packages.

### Step 4: Play the Game

To start playing, open the frontend URL in your browser (e.g., http://localhost:5176/).
