import Chessboard from "chessboardjsx";
import { useState } from "react";
import getQueryVariable from "./services/queryParser";
import wsfunctions from "./services/wsservice";
import {converters} from 'fen-reader'

const App = (props) =>{
  const [boardPos, setBoardPos] = useState("rnbqkbnr/pppppppp/11111111/11111111/11111111/11111111/PPPPPPPP/RNBQKBNR w KQkq - 0 1")


  const side = getQueryVariable("s") ? getQueryVariable("s") : "white" 

  wsfunctions.connectGame({setBoardPos, s:side})


  return (
    <div>
      <Chessboard position={boardPos} orientation={side}
      onDrop={({sourceSquare, targetSquare, piece}) => {
        

        if (sourceSquare === targetSquare) {
          return
        }
        console.log(`${piece} moving from ${sourceSquare} to ${targetSquare}`)

        {
          const board = converters.fen2json(boardPos.split(" ")[0])
          //console.log(board)
          let newBoard = {}
          for(let key of Object.keys(board)){
              newBoard[key] = board[key].charAt(1) + String(board[key].charAt(0)).toUpperCase();
          }
          delete newBoard[sourceSquare]
          newBoard[targetSquare] = piece
          setBoardPos(newBoard)
        }
        wsfunctions.sendMove(piece, sourceSquare, targetSquare)

        
       

        

      }}/>
      {/*<button onClick={() => wsfunctions.restartGame({})}>Restart</button>*/}
    </div>
  )
}

export default App;
