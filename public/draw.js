(function()
{
	var userName = null;
	while (!userName)
	{
		userName = prompt("Enter your username");
	}

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
	var currentDrawer = "";

	var eraseCheck = document.getElementById('erase');
	var clearBtn = document.getElementById('clearBtn');

	var textbox = document.getElementById('guessText');
	var guessBtn = document.getElementById('guessBtn');

	var drawerView = document.getElementById('drawerView');
	var guesserView = document.getElementById('guesserView');

	var guesses = document.getElementById('guesses');

	var leaderboard = document.getElementById('leaderboard');

	var currentDrawerView = document.getElementById('currentDrawer');
	var currentWordView = document.getElementById('currentWord');
	var messageView = document.getElementById('msg');

	var ws = new WebSocket("ws://" + window.location.host + "/socket");

	/********************* GUESSER FUNCTIONS *********************/


	function guess()
	{
		updateGuesses();
		if (isDrawer)
				return;

		var guessPacket = {};
		guessPacket.Type = "guess";
		guessPacket.Data = textbox.value;
		textbox.value = '';

		setTimeout(function() {
			sendToServer(guessPacket);
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
			doDraw(mouseX, mouseY, prevMouseX, prevMouseY, color);
			drawnPoints.push({
				'PrevX' : prevMouseX,
				'PrevY' : prevMouseY,
				'X' : mouseX,
				'Y' : mouseY,
				'Color' : color
			});
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

	function toggleCursor(value) {
		if (!isDrawer) {
			return;
		}
		eraseCheck.checked = value;
		if (value) {
			canvas.className = 'eraser';
		} else {
			canvas.className = '';
		}
	}

	function setWord()
	{
		var packet = {};
		packet.Type = "word";
		packet.Data = null;
		while (!packet.Data)
		{
			packet.Data = prompt("Choose the next word");
		}
		currentWordView.innerHTML = "The word is: " + packet.Data;
		currentWordView.display = "block";
		sendToServer(packet);
	}

	/********************* SHARED FUNCTIONS *********************/

	function displayMessage(msg)
	{
		messageView.innerHTML = msg
		messageView.display = "block";
		setTimeout(function()
		{
			messageView.innerHTML = null;
			messageView.display = "none";
		}, 3500);
	}

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

	function doDraw(xCoord, yCoord, prevXCoord, prevYCoord, currColor)
	{
		if (currColor != "white") {
			// drawing
			ctx.strokeStyle = currColor;
			ctx.beginPath();
			ctx.moveTo(prevXCoord, prevYCoord);
			ctx.lineTo(xCoord, yCoord);
			ctx.stroke();
			ctx.closePath();
		} else {
			// erasing
			ctx.fillStyle = currColor;
			ctx.fillRect(xCoord - 4, yCoord - 4, 8, 8);
		}
	}

	// handle message from server
	var read = function(event)
	{
		console.log("DELAY: " + delay);
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
					// TODO: parse and dispaly leaderboard, manually add ourselves
					// since the server recvs our name afterwards
					updateLeaderboard(parsed.Leaderboard);
					if (isDrawer)
					{
						currentDrawer = userName;
						setWord();
					}
					else
					{
						currentDrawer = parsed.Data;
						updateDraw(parsed);
					}

					currentDrawerView.innerHTML = "Current Drawer is " + currentDrawer;
					sendName();
					sendAck();
					displayMessage("Welcome " + userName + "!");
					break;
				case "badName":
					userName = prompt("Enter your username");
					sendName();
					sendAck();
					break;

				case "draw":
				    if (!isDrawer) {
						updateDraw(parsed);
					}
					sendAck();
					break;

				case "clear":
					clear();
					sendAck();
					break;

				case "drawerQuit":
					updateLeaderboard(parsed.Leaderboard);
					var oldDrawer = currentDrawer;
					currentDrawer = parsed.Data;
					currentDrawerView.innerHTML = "Current Drawer is " + currentDrawer;
					isDrawer = parsed.IsDrawer;
					toggleView();
					clear();

					if (isDrawer)
					{
						displayMessage(oldDrawer +
							" has quit, you are now drawing! The last word was " + parsed.LastWord);
						setWord();
					}
					else
					{
						sendAck();
						displayMessage(oldDrawer + " has quit, " + currentDrawer +
							" is now drawing. The last word was " + parsed.LastWord);
					}
					break;

				case "otherQuit":
					updateLeaderboard(parsed.Leaderboard);
					sendAck();
					displayMessage(parsed.Data + " has quit");
					break;

				case "next":
					// TODO: update drawer username here, also update the
					// new drawer's score
					updateLeaderboard(parsed.Leaderboard);
					currentWordView.innerHTML = null;
					currentWordView.display = "none";
					clear();
					clearGuesses();
					isDrawer = parsed.IsDrawer;
					currentDrawer = parsed.Data;
					updateDrawer(currentDrawer);
					toggleCursor(false);
					toggleView();

					if (isDrawer)
					{
						displayMessage("Correct guess! You are now the drawer");
						setWord();
					}

					else
					{
						sendAck();
						displayMessage("The last word was: " + parsed.LastWord +
							". " + currentDrawer + " is now drawing");
					}
					break;
				case "leaderboard":
					updateLeaderboard(parsed.Leaderboard);
					break;
			}
		}, delay);
	}

	// receive message from server
	ws.onmessage = read;

	function updateDraw(parsed)
	{
		if (!isDrawer && parsed.Board != null && parsed.Board.length != 0)
		{
			saveColor = color;
			color = parsed.Color;
			for (var i = 0; i < parsed.Board.length; i++)
			{
				doDraw(parsed.Board[i].x, parsed.Board[i].y,
					parsed.Board[i].prevX, parsed.Board[i].prevY,
					parsed.Board[i].color);
				// console.log("draw at: " + parsed.Board[i].x + " " +  parsed.Board[i].y);
				// ctx.fillRect(parsed.Board[i].x, parsed.Board[i].y, 1, 1);
			}
			color = saveColor;
		}
	}

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

	function updateLeaderboard(scores) {
		leaderboard.innerHTML = '';
		for (var user in scores) {
			leaderboard.innerHTML = leaderboard.innerHTML + '<div>' + user + ": " + scores[user] + '</div>';
		}
	}

	function updateDrawer(user) {
		currentDrawerView.innerHTML = "Current Drawer is " + user;
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
		toggleCursor(eraseCheck.checked);
	});

}());
