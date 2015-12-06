(function()
	{
		// client identifier
//		var guid = getId();
		var canvas = document.getElementById('canvas');
		var ctx = canvas.getContext("2d");

		var color = "black";
		var saveColor = "black";

		var mouseX = 0;
		var mouseY = 0;
		var prevMouseX = 0;
		var prevMouseY = 0;

		var isDrawer = false;

		var eraseCheck = document.getElementById('erase');
		var clearBtn = document.getElementById('clearBtn');

		var textbox = document.getElementById('guessText');
		var guessBtn = document.getElementById('guessBtn');

		var ws = new WebSocket("ws://localhost:7777/socket");

		var read = function(event)
		{
			var parsed = JSON.parse(event.data);
			isDrawer = parsed.IsDrawer;
			console.log(parsed);

			if (!isDrawer)
			{
				saveColor = color;
				color = parsed.Color;
				for (var i = 1; i < parsed.Board.length; i++)
				{
					// console.log("draw at: " + parsed.Board[i].x + " " +  parsed.Board[i].y);
					// ctx.fillRect(parsed.Board[i].x, parsed.Board[i].y, 1, 1);
					ctx.strokeStyle = color;
					ctx.beginPath();
					ctx.moveTo(parsed.Board[i-1].x, parsed.Board[i-1].y)
					ctx.lineTo(parsed.Board[i].x, parsed.Board[i].y);
					ctx.stroke();
					ctx.closePath();
				}
				color = saveColor;
			}
		}

		// start two intervals:
		// every 10 ms:
			// if drawer: send entire drawed range
			// for both: read godpacket from server

		ws.onmessage = read;

		var drawnPoints = [];

		setInterval(function()
		{
			if (isDrawer)
			{
				flush();
			}
		}, 200);

		function flush()
		{
			if (drawnPoints.length == 0)
			{
				return;
			}

			var packet = {};
			packet.Board = drawnPoints;
			packet.Color = color;
			console.log(JSON.stringify(packet));
			sendToServer(JSON.stringify(packet));
			drawnPoints = [];
		}

		window.addEventListener('beforeunload', function (e) {
			quit();
		}, false);

		textbox.addEventListener('keypress', function(e) {
			if (e.keyCode === 13) {
				guess();
				textbox.value = '';
			}
		});

		canvas.addEventListener('mousemove', function(e) {
			mousePos(e);
		}, false);

		canvas.addEventListener('mousedown', function(e) {
			canvas.addEventListener('mousemove', draw, false);
		}, false);

		canvas.addEventListener('mouseup', function(e) {
			canvas.removeEventListener('mousemove', draw, false);
		}, false);

		clearBtn.addEventListener('click', function(e) {
			clear();
		});

		guessBtn.addEventListener('click', function(e) {
			guess();
		});

		eraseCheck.addEventListener('click', function(e) {
			toggleEraser();
		});


/*		function getId() {
			 getFromServer(""). ;
		}*/

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

			ctx.strokeStyle = color;
			ctx.beginPath();
			ctx.moveTo(prevMouseX, prevMouseY)
				ctx.lineTo(mouseX, mouseY);
			ctx.stroke();
			ctx.closePath();
			var location =	{'X' : mouseX, 'Y' : mouseY} // , 'c': color} DO THIS LATER
			// sendToServer(JSON.stringify(location));
			drawnPoints.push(location);
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

		function clear()
		{
			ctx.clearRect(0, 0, canvas.width, canvas.height);
		}

		function quit()
		{
			sendToServer("quit");
			ws.close();
		}

		var guess = function()
		{
			// if (isDrawer)
				//   return;

			sendToServer(JSON.stringify(textbox.value));
		}

		function sendToServer(data)
		{
			ws.send(data);
		}

	}());
