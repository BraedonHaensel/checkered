export const PlayerColor = {
  RED: 'red',
  BLACK: 'black',
}
export type PlayerColor = (typeof PlayerColor)[keyof typeof PlayerColor]
