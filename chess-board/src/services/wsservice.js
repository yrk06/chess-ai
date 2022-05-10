import {converters} from 'fen-reader'

let ws;

let wsclose = true;

let gameover = false;

const connectGame = (state) => {
    if ( (ws == null || wsclose) && !gameover) {
        ws = new WebSocket(`${window.location.protocol === "https:" ? 'wss': 'ws'}://localhost:8080/${state.t}`)
        
        wsclose = false
    }
    ws.onopen = () => {
        ws.send(state.s)
    }
    ws.onclose = () => {
        wsclose = true
    }
    ws.onmessage = (data) => {
        const {setBoardPos, setEval} = state
        console.log(data.data)

        if (data.data.startsWith("eval")) {
            const limite = 2000
            const value = Math.max(Math.min(parseFloat(data.data.split(" ")[1]), limite),-limite)

            console.log(parseFloat(data.data.split(" ")[1]))

            console.log( ( (value/limite)/2.0 + 0.5 ) * 100.0 );
            setEval( ( (value/limite)/2.0 + 0.5 ) * 100.0 )

        } else {
            const board = converters.fen2json(data.data.split(" ")[0])
            //console.log(board)
            let newBoard = {}
            for(let key of Object.keys(board)){
                newBoard[key] = board[key].charAt(1) + String(board[key].charAt(0)).toUpperCase();
            }
            if (data.data === "Checkmate") {
                gameover = true
            }
            setBoardPos(data.data)
        }

        
        
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
