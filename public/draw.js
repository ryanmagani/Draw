window.onload = (function()
{
  var canvas = document.getElementsByTagName('canvas')[0];
  var ctx = canvas.getContext("2d");
  // ctx.fillStyle = "solid";
  // ctx.strokeStyle = "#000000";
  ctx.lineWidth = 30;
  ctx.beginPath(); // not needed apparently
  ctx.moveTo(0, 0);
  ctx.lineTo(100, 100);
  ctx.stroke();
  ctx.closePath();
});
