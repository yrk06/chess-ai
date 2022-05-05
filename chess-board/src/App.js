import Chessboard from "chessboardjsx";
import { useState } from "react";
import getQueryVariable from "./services/queryParser";
import wsfunctions from "./services/wsservice";
import { converters } from 'fen-reader'
import { Col, Container, Progress, Row, } from 'reactstrap'

import logo from './images/XGC.png';

const App = (props) => {
  const [boardPos, setBoardPos] = useState("rnbqkbnr/pppppppp/11111111/11111111/11111111/11111111/PPPPPPPP/RNBQKBNR w KQkq - 0 1")

  const [evalScore, setEval] = useState(50)


  const side = getQueryVariable("s") ? getQueryVariable("s") : "white"
  const ai = getQueryVariable("t") ? getQueryVariable("t") : "echo"

  wsfunctions.connectGame({ setBoardPos, s: side, t: ai, setEval })


  return (
    <Container>
      <Row className="mx-auto mt-3" style={{ width: "560px" }}>
        <Col> 
        {/*rgb(66, 245, 242) rgb(235, 52, 155)*/}
          <img className="" src={logo} width="100px" style={{backgroundColor: 'rgb(66, 245, 242)' ,marginLeft: `${(560-100)/2}px`,border: '2px solid black', borderRadius:'20px'}} alt="XGC" />
          <h1 className="text-center" style={{color: 'rgb(66, 245, 242)'}}> eXtreme Go Chess </h1>
        </Col>
        
      </Row>

      <Row className="mx-auto" style={{ width: "560px" }}>
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

      <Row className="mx-auto mt-2" style={{ width: "560px" }}>
        <Col >
          <div className="progress" style={{ backgroundColor:'rgb(41,41,41)',height: "60px", width: "560px",padding: '5px' }}>

            
            <div className="progress-bar" role="progressbar" style={{width: `${Math.max(Math.min(evalScore,50),0)}%`,backgroundColor: 'rgb(41,41,41)'}}></div>
            <div className="progress-bar progress-bar-striped progress-bar-animated" role="progressbar" style={{width: `${50-Math.max(Math.min(evalScore,50),0)}%`,backgroundColor: 'rgb(237,65,123)'}}></div>
            
            <div className="progress-bar progress-bar-striped progress-bar-animated" role="progressbar" style={{width: `${Math.max(Math.min(evalScore-50,50),0)}%`,backgroundColor: 'rgb(66, 245, 242)'}}></div>
            <div className="progress-bar" role="progressbar" style={{width: `${50-Math.max(Math.min(evalScore-50,50),0)}%`,backgroundColor: 'rgb(41,41,41)'}}></div>
          
          </div>
        </Col>
      </Row>

    </Container>
  )
}

export default App;
