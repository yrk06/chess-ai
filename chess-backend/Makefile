white:
	rm -rf build/static
	cd ../chess-board; npm run build; mv build/ ../chess-backend/build/static/;
	go build -ldflags="-H windowsgui" -o build/white.exe

black:
	rm -rf build/static
	cd ../chess-board; npm run build; mv build/ ../chess-backend/build/static/;
	go build -ldflags="-H windowsgui" -o build/black.exe

server:
	rm -rf build/static
	cd ../chess-board; npm run build; mv build/ ../chess-backend/build/static/;
	go build -ldflags="-H windowsgui" -o build/server.exe

frontend:
	rm -rf build/static
	cd ../chess-board; npm run build; mv build/ ../chess-backend/build/static/;

