import Chessboard from "chessboardjsx";
import { useState } from "react";
import getQueryVariable from "./services/queryParser";
import wsfunctions from "./services/wsservice";
import { converters } from 'fen-reader'
import { Col, Container, Progress, Row } from 'reactstrap'

import logo from './images/XGC.png';

const App = (props) => {
  const [boardPos, setBoardPos] = useState("rnbqkbnr/pppppppp/11111111/11111111/11111111/11111111/PPPPPPPP/RNBQKBNR w KQkq - 0 1")

  const [evalScore, setEval] = useState(50)


  const side = getQueryVariable("s") ? getQueryVariable("s") : "white"
  const ai = getQueryVariable("t") ? getQueryVariable("t") : "echo"

  wsfunctions.connectGame({ setBoardPos, s: side, t: ai, setEval })


  return (
    <Container>
      <h1 className="text-center"> eXtreme Go Chess <img src={logo} width="200px" alt="XGC"/></h1>
      <Row className="mx-auto" style={{width:"560px"}}>
        <Col>
          <Chessboard className="mx-auto" position={boardPos} sparePieces={false} /*orientation={side}*/
            darkSquareStyle={{ backgroundColor: "rgb(41, 41, 41)" }}
            lightSquareStyle={{ backgroundColor: "rgb(66, 245, 242)" }}
            onDrop={({ sourceSquare, targetSquare, piece }) => {


              if (sourceSquare === targetSquare) {
                return
              }
              console.log(`${piece} moving from ${sourceSquare} to ${targetSquare}`)

              {
                const board = converters.fen2json(boardPos.split(" ")[0])
                //console.log(board)
                let newBoard = {}
                for (let key of Object.keys(board)) {
                  newBoard[key] = board[key].charAt(1) + String(board[key].charAt(0)).toUpperCase();
                }
                delete newBoard[sourceSquare]
                newBoard[targetSquare] = piece
                setBoardPos(newBoard)
              }
              wsfunctions.sendMove(piece, sourceSquare, targetSquare)






            }} />
        </Col>
      </Row>

      <Row className="mx-auto" style={{ width: "560px" }}>
        <Col style={{ width: "720px" }}>
          <Progress style={{height:"50px", width: "560px"}} value={evalScore} animated={true} color={(evalScore > 50) ? 'success' : 'danger'} />
        </Col>
      </Row>

    </Container>
  )
}

export default App;
