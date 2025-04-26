const express = require('express');
const path = require('path');
const bodyParser = require('body-parser');
const methodOverride = require('method-override');
require('dotenv').config();

const eventRoutes = require('./routes'); // Original routes for event management
const bettingRoutes = require('./bettingRoutes'); // <-- Import the new routes

const app = express();
const port = process.env.PORT || 3000;

// View engine setup
app.set('views', path.join(__dirname, 'views'));
app.set('view engine', 'ejs');

// Middleware (ensure these are before routes)
app.use(express.static(path.join(__dirname, 'public')));
app.use(bodyParser.urlencoded({ extended: true }));
app.use(bodyParser.json());
app.use(methodOverride(function (req, res) {
  // (Keep the existing methodOverride logic)
  if (req.query && typeof req.query === 'object' && '_method' in req.query) {
    var method = req.query._method;
    delete req.query._method;
    return method;
  }
   if (req.body && typeof req.body === 'object' && '_method' in req.body) {
    var method = req.body._method
    delete req.body._method
    return method
  }
}));

// Use Routes
app.use('/', eventRoutes); // Keep original event management at root
app.use('/betting', bettingRoutes); // <-- Mount betting routes under /betting

// Catch 404 and forward to error handler
app.use(function(req, res, next) {
  // (Keep existing 404 handler)
  const err = new Error('Not Found');
  err.status = 404;
  res.status(404).render('error', { message: 'Page Not Found', status: 404});
});

// Basic error handler
app.use(function(err, req, res, next) {
  // (Keep existing error handler)
  console.error("Global Error Handler Caught:", err); // Add logging
  res.status(err.status || 500);
  res.render('error', {
    message: err.message || 'An internal server error occurred.',
    status: err.status || 500,
  });
});

// Start server
app.listen(port, () => {
    console.log(`Server running on http://localhost:${port}`);
    console.log(`Event API configured at: ${process.env.API_BASE_URL}`);
    console.log(`Betting API configured at: ${process.env.BETTING_API_BASE_URL}`); 
});