// app.js
const express = require('express');
const path = require('path');
const bodyParser = require('body-parser');
const methodOverride = require('method-override');
require('dotenv').config(); // Ensure env vars are loaded

const routes = require('./routes'); // Import the routes

const app = express();
const port = process.env.PORT || 3000;

// View engine setup
app.set('views', path.join(__dirname, 'views'));
app.set('view engine', 'ejs');

// Middleware
app.use(express.static(path.join(__dirname, 'public'))); // Serve static files (CSS)
app.use(bodyParser.urlencoded({ extended: true })); // Parse URL-encoded bodies (form data)
app.use(bodyParser.json()); // Parse JSON bodies (though less common for server-rendered forms)

// Method Override: Allows forms to use PUT, DELETE, PATCH etc.
// Looks for _method query parameter or hidden form field
app.use(methodOverride(function (req, res) {
  if (req.query && typeof req.query === 'object' && '_method' in req.query) {
    // Look in url query parameters for _method
    var method = req.query._method;
    delete req.query._method;
    return method;
  }
   if (req.body && typeof req.body === 'object' && '_method' in req.body) {
    // Look in urlencoded POST bodies and delete it
    var method = req.body._method
    delete req.body._method
    return method
  }
}));

// Use Routes
app.use('/', routes);

// Catch 404 and forward to error handler (optional, basic)
app.use(function(req, res, next) {
  const err = new Error('Not Found');
  err.status = 404;
  res.status(404).render('error', { message: 'Page Not Found', status: 404});
});

// Basic error handler (optional, basic)
app.use(function(err, req, res, next) {
  // Render the error page
  res.status(err.status || 500);
  res.render('error', {
    message: err.message,
    status: err.status || 500,
    // Only provide stack in development (optional)
    // error: req.app.get('env') === 'development' ? err : {}
  });
});


// Start server
app.listen(port, () => {
    console.log(`Server running on http://localhost:${port}`);
    console.log(`Connecting to API at: ${process.env.API_BASE_URL}`);
});