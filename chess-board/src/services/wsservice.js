import {converters} from 'fen-reader'

let ws;

let wsclose = true;

const connectGame = (state) => {
    if (ws == null || wsclose) {
        ws = new WebSocket(`ws://localhost:8080/echo`)
        ws.onopen = () => {
            ws.send("hi")
        }
        ws.onclose = () => {
            wsclose = true
        }
        ws.onmessage = (data) => {
            const {setBoardPos} = state
            const board = converters.fen2json(data.data.split(" ")[0])
            console.log(board)
            for(let key of Object.keys(board)){
                board[key] = board[key].charAt(1) + String(board[key].charAt(0)).toUpperCase();
            }
            console.log(board)
            setBoardPos(board)
        }
        wsclose = false
    }
    
}

const restartGame = (state) => {
    ws.close()
    ws = null
    connectGame(state)
}

const sendMove = (piece, startingSquare, targetSquare) => {
    ws.send(`${piece}-${startingSquare}-${targetSquare}`)
}

const wsfunctions = {connectGame, sendMove, restartGame}

export default wsfunctions;
