// routes.js
const express = require('express');
const apiService = require('./services/apiService');
const router = express.Router();

// Middleware to handle flash messages (simple version)
router.use((req, res, next) => {
  res.locals.message = req.query.message; // Get message from query param
  res.locals.messageType = req.query.type || 'info'; // 'info', 'success', 'error'
  next();
});

// Helper function to render error page
const renderError = (res, err, defaultMessage = 'An unexpected error occurred.') => {
    const message = err.message || defaultMessage;
    const status = err.status || 500;
    res.status(status).render('error', { message, status });
};

// --- Event Routes ---

// GET / - List all events
router.get('/', async (req, res) => {
    try {
        const events = await apiService.getAllEvents();
        // Sort events, e.g., by start date descending
        events.sort((a, b) => new Date(b.eventStartDate) - new Date(a.eventStartDate));
        res.render('index', { events });
    } catch (err) {
        console.error("Error fetching events:", err);
        renderError(res, err, 'Could not fetch events from the API.');
    }
});

// GET /events/new - Show form to create a new event
router.get('/events/new', (req, res) => {
    res.render('create-event');
});

// POST /events - Create a new event
router.post('/events', async (req, res) => {
    try {
        const newEvent = await apiService.createEvent(req.body);
        res.redirect(`/?message=Event '${newEvent.eventName}' created successfully&type=success`);
    } catch (err) {
        console.error("Error creating event:", err);
        // If validation error, might want to re-render form with error messages
        renderError(res, err, 'Failed to create event.');
    }
});

// GET /events/:id - Show event details
router.get('/events/:id', async (req, res) => {
    try {
        const eventId = req.params.id;
        const event = await apiService.getEventDetails(eventId); // Use details endpoint
        if (!event) {
            return renderError(res, { status: 404, message: `Event with ID ${eventId} not found.` });
        }
        res.render('event-detail', { event });
    } catch (err) {
        console.error(`Error fetching event details for ${req.params.id}:`, err);
        renderError(res, err, 'Could not fetch event details.');
    }
});

// POST /events/:id/end - End an event
router.post('/events/:id/end', async (req, res) => {
    const eventId = req.params.id;
    try {
        const result = await apiService.endEvent(eventId);
        res.redirect(`/events/${eventId}?message=Event ended. Result: ${result.result}&type=success`);
    } catch (err) {
        console.error(`Error ending event ${eventId}:`, err);
        res.redirect(`/events/${eventId}?message=Error ending event: ${err.message}&type=error`);
    }
});

// POST /events/:id/cancel - Cancel an event
router.post('/events/:id/cancel', async (req, res) => {
    const eventId = req.params.id;
    try {
        await apiService.cancelEvent(eventId);
        res.redirect(`/events/${eventId}?message=Event successfully cancelled.&type=success`);
    } catch (err) {
        console.error(`Error cancelling event ${eventId}:`, err);
        res.redirect(`/events/${eventId}?message=Error cancelling event: ${err.message}&type=error`);
    }
});

// POST /events/:id/subscribe - Subscribe to an event
router.post('/events/:id/subscribe', async (req, res) => {
    const eventId = req.params.id;
    const { callbackUrl } = req.body;
    if (!callbackUrl) {
        return res.redirect(`/events/${eventId}?message=Callback URL is required.&type=error`);
    }
    try {
        await apiService.subscribeToEvent(eventId, callbackUrl);
        res.redirect(`/events/${eventId}?message=Successfully subscribed with URL: ${callbackUrl}&type=success`);
    } catch (err) {
        console.error(`Error subscribing to event ${eventId}:`, err);
        res.redirect(`/events/${eventId}?message=Subscription failed: ${err.message}&type=error`);
    }
});

// DELETE /events/:id - Delete an event
router.delete('/events/:id', async (req, res) => {
    const eventId = req.params.id;
    try {
        await apiService.deleteEvent(eventId);
        res.redirect('/?message=Event successfully deleted.&type=success');
    } catch (err) {
        console.error(`Error deleting event ${eventId}:`, err);
        // Redirect back to home page with error as event page won't exist
         res.redirect(`/?message=Error deleting event: ${err.message}&type=error`);
    }
});

// --- Event Detail Routes ---

// GET /events/:id/details/new - Show form to add a new round/detail
router.get('/events/:id/details/new', (req, res) => {
    res.render('create-detail', { eventId: req.params.id });
});

// POST /events/:id/details - Create a new event detail/round
router.post('/events/:id/details', async (req, res) => {
    const eventId = req.params.id;
    try {
        await apiService.createEventDetail(eventId, req.body);
        res.redirect(`/events/${eventId}?message=New round/detail added successfully.&type=success`);
    } catch (err) {
        console.error(`Error creating event detail for ${eventId}:`, err);
        res.redirect(`/events/${eventId}?message=Error adding round: ${err.message}&type=error`);
    }
});

// PATCH /events/:id/details - Update an existing event detail/round (via form POST + method-override)
// Note: This requires more complex form handling on the event-detail page or a separate edit page.
// The C# API uses PATCH for updating *a* detail based on RoundNumber in the body.
// This example assumes you might POST to this endpoint with _method=PATCH
router.patch('/events/:id/details', async (req, res) => {
    const eventId = req.params.id;
     try {
        // req.body should contain roundNumber, homeTeamScore, awayTeamScore
        if (!req.body.roundNumber) {
            throw new Error("Round number is required for update.");
        }
        await apiService.updateEventDetail(eventId, req.body);
        res.redirect(`/events/${eventId}?message=Round ${req.body.roundNumber} updated successfully.&type=success`);
    } catch (err) {
        console.error(`Error updating event detail for ${eventId}:`, err);
        res.redirect(`/events/${eventId}?message=Error updating round ${req.body.roundNumber || ''}: ${err.message}&type=error`);
    }
});


module.exports = router;