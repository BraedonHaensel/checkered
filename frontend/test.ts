interface Message<type> {
    type: type,
}

interface WinMessage extends Message<"win"> {
    winner: string,
}
