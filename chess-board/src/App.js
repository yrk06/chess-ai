import Chessboard from "chessboardjsx";
import { useState } from "react";

const pos2obj = (pos) => {
  return {
    x: pos.charCodeAt(0) - 97,
    y: Number(pos[1])
  }
}

const obj2pos = (pos) => {
  return String.fromCharCode(pos.x+97)+String(pos.y)
}

const pawnMove = (team,pos,board) => {
  const p = pos2obj(pos)
  const moves = []
  let np = {...p}
  if (team == "w"){
    np.y += 1
  }
  np.x += 1
  if (obj2pos(np) in board)
  {
    moves.push(obj2pos(np))
  }
  np.x -= 2
  if (obj2pos(np) in board)
  {
    moves.push(obj2pos(np))
  }


  np = {...p}
  if(team == 'w') {
    np.y += 1
  }
  if (obj2pos(np) in board)
  {
    return moves
  }
  moves.push(obj2pos(np))
  if(p.y == 2){
    np = {...p}
    if(team == 'w') {
      np.y += 2
    }
    if (obj2pos(np) in board)
    {
      return moves
    }
  }
  moves.push(obj2pos(np))
  return moves
}

const checkMove = (piece,pos,board) => {
  const team = piece[0]
  const p = piece[1]

  switch(p) {
    case 'P': {
      return pawnMove(team,pos,board);
    }
    default: {
      return []
    }
  }
}

const startRow = ['R','N','B','Q','K','B','N','R']
let firstBoardState = {};
for(let i = 0; i < startRow.length; i++){
  firstBoardState[`${String.fromCharCode(i+97)}1`] = 'w'+startRow[i]
}
for(let i = 0; i < startRow.length; i++){
  firstBoardState[`${String.fromCharCode(i+97)}2`] = 'wP'
}
for(let i = 0; i < startRow.length; i++){
  firstBoardState[`${String.fromCharCode(i+97)}8`] = 'b'+startRow[i]
}
for(let i = 0; i < startRow.length; i++){
  firstBoardState[`${String.fromCharCode(i+97)}7`] = 'bP'
}




const App = () =>{
  const [boardPos, setBoardPos] = useState(firstBoardState
    )
  return (
    <div>
      <Chessboard position={boardPos}
      onDrop={({sourceSquare, targetSquare, piece}) => {
        console.log(`${piece} moving from ${sourceSquare} to ${targetSquare}`)
        
        if(!checkMove(piece,sourceSquare,boardPos).includes(targetSquare))
        {
          console.log("invalid");
        }

        const newPos = {...boardPos}

        newPos[sourceSquare] = undefined
        newPos[targetSquare] = piece

        if (newPos[sourceSquare] == undefined)
        {
          delete newPos[sourceSquare]
        }
        
        
        setBoardPos(newPos)

      }}/>
    </div>
  )
}

export default App;
