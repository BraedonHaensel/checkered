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

Use the `-addr` flag to change the server's address/port.
Use the `-ns` flag to change the Name Server's fully qualified URL

## Name Server

Run the Name Server using the following:

```sh
cd name-server && go run .
```

Use the `-addr` flag to change the Name Server's address/port.
