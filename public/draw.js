(function()
{
	var userName = prompt("Enter your username");

	var artificialDelay = 0;
	var delay = 0;

	var canvas = document.getElementById('canvas');
	var ctx = canvas.getContext("2d");

	var color = "black";
	var saveColor = "black";

	// buffer for drawn points to send to server
	var drawnPoints = [];

	// locations of draw
	var mouseX = 0;
	var mouseY = 0;
	var prevMouseX = 0;
	var prevMouseY = 0;

	var isDrawer = false;
	var currentDrawer = ""; // TODO: use this field

	var eraseCheck = document.getElementById('erase');
	var clearBtn = document.getElementById('clearBtn');

	var textbox = document.getElementById('guessText');
	var guessBtn = document.getElementById('guessBtn');

	var drawerView = document.getElementById('drawerView');
	var guesserView = document.getElementById('guesserView');

	var guesses = document.getElementById('guesses');

	var leaderboard = document.getElementById('leaderboard');

	var ws = new WebSocket("ws://localhost:7777/socket");

	/********************* GUESSER FUNCTIONS *********************/


	function guess()
	{
		updateGuesses();

		setTimeout(function() {
			 if (isDrawer)
				return;

			var guessPacket = {};
			guessPacket.Type = "guess";
			guessPacket.Data = textbox.value;

			sendToServer(guessPacket);
			textbox.value = '';
		}, delay);
	}

	function updateGuesses() {
		guesses.innerHTML = "<div>" + textbox.value + "</div>" + guesses.innerHTML;
	}

	function clearGuesses() {
		guesses.innerHTML = "";
	}

	/********************* DRAWER FUNCTIONS *********************/


	// flush buffered drawing to server
	setInterval(function()
	{
		if (isDrawer)
		{
			flush();
		}
	}, 200);

	function mousePos(e)
	{
		prevMouseX = mouseX;
		prevMouseY = mouseY;
		mouseX = e.clientX - canvas.offsetLeft;
		mouseY = e.clientY - canvas.offsetTop;
	}

	function draw()
	{
		if (!isDrawer)
			return;

		if (mouseX > 0 && mouseX < 400 && mouseY > 0 && mouseY < 400) {
			doDraw(mouseX, mouseY, prevMouseX, prevMouseY);
			drawnPoints.push({'X' : mouseX, 'Y' : mouseY});
		}
	}

	function flush()
	{
		if (drawnPoints.length == 0)
		{
			return;
		}

		var packet = {};
		packet.Type = "draw";
		packet.Board = drawnPoints;
		packet.Color = color;
		console.log(JSON.stringify(packet));
		sendToServer(packet);
		drawnPoints = [];
	}

	function toggleEraser()
	{
		flush();

		if (eraseCheck.checked)
		{
			saveColor = color;
			color = "white";
		}

		else
		{
			color = saveColor;
		}
	}


	/********************* SHARED FUNCTIONS *********************/

	function sendAck()
	{
		var packet = {};
		packet.Type = "ack";
		packet.Date = Date.now();

		sendToServer(packet);
	}

	function sendName()
	{
		var packet = {};
		packet.Type = "name";
		packet.Data = userName;
		sendToServer(packet);
	}

	function doDraw(xCoord, yCoord, prevXCoord, prevYCoord)
	{
		ctx.strokeStyle = color;
		ctx.beginPath();
		ctx.moveTo(prevXCoord, prevYCoord)
			ctx.lineTo(xCoord, yCoord);
		ctx.stroke();
		ctx.closePath();
	}

	// handle message from server
	var read = function(event)
	{
		setTimeout(function() {

		
			var parsed = JSON.parse(event.data);
			if (parsed.Delay)
			{
				delay = parsed.Delay;
			}
	console.log(parsed);
			switch (parsed.Type)
			{
				case "init":
					isDrawer = parsed.IsDrawer;
					toggleView();
					// TODO: fill the whole board
					// TODO: parse and dispaly leaderboard, manually add ourselves
					// since the server recvs our name afterwards
					sendName();
					sendAck();
					break;
				case "badName":
					userName = prompt("Enter your username");
					sendName();
					sendAck();
					break;

				case "draw":
					if (!isDrawer && parsed.Board != null && parsed.Board.length != 0)
					{
						saveColor = color;
						color = parsed.Color;
						for (var i = 1; i < parsed.Board.length; i++)
						{
							doDraw(parsed.Board[i].x, parsed.Board[i].y,
								parsed.Board[i-1].x, parsed.Board[i-1].y);
							// console.log("draw at: " + parsed.Board[i].x + " " +  parsed.Board[i].y);
							// ctx.fillRect(parsed.Board[i].x, parsed.Board[i].y, 1, 1);
						}
						color = saveColor;
					}
					sendAck();
					break;

				case "clear":
					clear();
					sendAck();
					break;

				case "drawerQuit":
					removeFromLeaderboard(currentDrawer);
					currentDrawer = parsed.Data;
					isDrawer = parsed.IsDrawer;
					toggleView();
					clear();
					sendAck();
					break;

				case "otherQuit":
					removeFromLeaderboard(parsed.Data);
					sendAck();
					break;

				case "next":
					// TODO: update drawer username here, also update the
					// new drawer's score
					increaseScore(parsed.Data);
					clear();
					clearGuesses();
					isDrawer = parsed.IsDrawer;
					drawer = parsed.Data;
					toggleView();
					sendAck();
					break;
			}
		}, delay);
	}

	// receive message from server
	ws.onmessage = read;

	// Clears board, called by server if we are a guesser
	function clear()
	{
		ctx.clearRect(0, 0, canvas.width, canvas.height);
		var clearPacket = {};
		clearPacket.Type = "clear";
		if (isDrawer)
		{
			sendToServer(clearPacket);
		}
	}

	// quit game
	function quit()
	{
		var quitPacket = {};
		quitPacket.Type = "quit";
		sendToServer(quitPacket);
		ws.close();
	}

	function removeFromLeaderboard(user)
	{
		var person = document.getElementById(user);
		if (person) {
			person.parentElement.removeChild(person);
		}
	}

	function increaseScore(user)
	{
		var person = document.getElementById(user);
		if (person) {
			var score = person.getAttribute('score');
			score++;
			person.innerHTML = user + ": " + score;
		} else {
			person = document.createElement("div");
			person.id = user;
			person.setAttribute('score', 1);
			person.innerHTML = user + ": " + 1;
			leaderboard.appendChild(person);
		}
	}

	// send data to server
	function sendToServer(data)
	{
		ws.send(JSON.stringify(data));
	}

	function toggleView() {
		if (isDrawer) {
			drawerView.style.display = "block";
			guesserView.style.display = "none";
		} else {
			drawerView.style.display = "none";
			guesserView.style.display = "block";
		}
	}

	/********************* EVENT LISTENERS *********************/


	window.onbeforeunload = function () {
		quit();
	};

	canvas.addEventListener('mousemove', function(e) {
		mousePos(e);
	}, false);

	canvas.addEventListener('mousedown', function(e) {
		canvas.addEventListener('mousemove', draw, false);
	}, false);

	canvas.addEventListener('mouseup', function(e) {
		canvas.removeEventListener('mousemove', draw, false);
	}, false);

	textbox.addEventListener('keypress', function(e) {
		if (e.keyCode === 13) {
			guess();
		}
	});

	clearBtn.addEventListener('click', function(e) {
		clear();
	});

	guessBtn.addEventListener('click', function(e) {
		guess();
	});

	eraseCheck.addEventListener('click', function(e) {
		toggleEraser();
	});

}());
