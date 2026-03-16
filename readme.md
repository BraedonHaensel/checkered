# Checkered

## Frontend 

Frontend is written using Vite/React.js. 

First install dependencies:
```sh
cd frontend && npm install
```
Then start the development server:
```sh
npm run dev
```

## Backend 
To run a matchmaker server
```sh
cd backend && make matchmaker
```
To run a game server
```sh
cd backend && make game-server
```

## Name Server
Run the Name Server using the following:
```sh
cd name-server && go run .
```

Use the `-addr` flag to change the Name Server's address/port.

Edit `game-servers.txt` and `matchmaking-servers.txt` to change the list of
active servers.
