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
Run the backend using the following:
```sh
cd backend && go run .
```

## Name Server
Run the Name Server using the following:
```sh
cd name-server && go run .
```

Use the `-addr` flag to change the Name Server's address/port.

Edit `game-servers.txt` and `matchmaking-servers.txt` to change the list of
active servers.
