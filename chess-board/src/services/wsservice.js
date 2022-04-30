import {converters} from 'fen-reader'

let ws;

let wsclose = true;

const connectGame = (state) => {
    if (ws == null || wsclose) {
        ws = new WebSocket(`ws://localhost:8080/echo`)
        
        wsclose = false
    }
    ws.onopen = () => {
        ws.send(state.s)
    }
    ws.onclose = () => {
        wsclose = true
    }
    ws.onmessage = (data) => {
        const {setBoardPos} = state
        console.log(data.data)
        const board = converters.fen2json(data.data.split(" ")[0])
        //console.log(board)
        let newBoard = {}
        for(let key of Object.keys(board)){
            newBoard[key] = board[key].charAt(1) + String(board[key].charAt(0)).toUpperCase();
        }
        setBoardPos(data.data)
        
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
