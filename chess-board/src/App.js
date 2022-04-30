import Chessboard from "chessboardjsx";
import { useState } from "react";
import wsfunctions from "./services/wsservice";

const App = () =>{
  const [boardPos, setBoardPos] = useState("rnbqkbnr/pppppppp/11111111/11111111/11111111/11111111/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
  
  wsfunctions.connectGame({setBoardPos})

  return (
    <div>
      <Chessboard position={boardPos}
      getPosition={ position => setBoardPos(position)}
      onDrop={({sourceSquare, targetSquare, piece}) => {
        

        if (sourceSquare == targetSquare) {
          return
        }

        console.log(boardPos)
        console.log(`${piece} moving from ${sourceSquare} to ${targetSquare}`)

        setTimeout( () => {
          wsfunctions.sendMove(piece, sourceSquare, targetSquare)
        },100)
       

        if ( typeof(boardPos) != "string" ){
          const newBoard = {...boardPos}
          delete newBoard[sourceSquare]
          newBoard[targetSquare] = piece
          setBoardPos(newBoard)
        }

      }}/>
      <button onClick={() => wsfunctions.restartGame({setBoardPos})}>Restart</button>
    </div>
  )
}

export default App;
