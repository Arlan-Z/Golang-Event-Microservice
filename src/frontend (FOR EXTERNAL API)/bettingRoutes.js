const express = require('express');
const bettingApiService = require('./services/bettingApiService'); // Use the new service
const router = express.Router();

// Helper function to render error page (can be reused or made global)
const renderError = (res, err, defaultMessage = 'An unexpected error occurred.') => {
    const message = err.message || defaultMessage;
    const status = err.status || 500;
    // Pass details if they exist from the API error
    const details = err.details ? JSON.stringify(err.details, null, 2) : null;
    console.error(`Rendering error page: Status ${status}, Message: ${message}`, err.details || '');
    res.status(status).render('error', { // Assuming you have views/error.ejs
         message: `${message}${details ? ` (Details: ${details})` : ''}`,
         status
        });
};

// Middleware for flash messages (same as before)
router.use((req, res, next) => {
  res.locals.message = req.query.message;
  res.locals.messageType = req.query.type || 'info';
  next();
});

// GET /betting/ - Display active events and betting forms
router.get('/', async (req, res) => {
    try {
        const events = await bettingApiService.getActiveEvents();
        // Optional: Check readiness/health here if you want to display status
        // const readiness = await bettingApiService.checkReadiness();
        res.render('betting-index', {
             events: events || [], // Ensure events is always an array
             // readinessStatus: readiness.status
             });
    } catch (err) {
        console.error("Error fetching active betting events:", err);
        renderError(res, err, 'Could not fetch active events from the Betting API.');
    }
});

// POST /betting/bets - Place a bet
router.post('/bets', async (req, res) => {
    const { userId, eventId, amount, predictedOutcome } = req.body;
    try {
        // Basic validation within the route too (though service has more)
        if (!userId || !eventId || !amount || !predictedOutcome) {
            throw { status: 400, message: 'Missing required fields for placing a bet.' };
        }
        const betResult = await bettingApiService.placeBet({ userId, eventId, amount, predictedOutcome });
        const message = encodeURIComponent(`Bet placed successfully! Bet ID: ${betResult.id}, Amount: ${betResult.amount}, Outcome: ${betResult.predictedOutcome}`);
        res.redirect(`/betting?message=${message}&type=success`);
    } catch (err) {
        console.error("Error placing bet:", err);
        const message = encodeURIComponent(`Failed to place bet: ${err.message}`);
        // Redirect back with error message
        res.redirect(`/betting?message=${message}&type=error`);
        // Alternatively, render an error page:
        // renderError(res, err, `Failed to place bet.`);
    }
});

// POST /betting/events/:id/finalize - Finalize an event
router.post('/events/:id/finalize', async (req, res) => {
    const eventId = req.params.id;
    const { result } = req.body; // Result comes from the form select
     try {
        if (!result) {
             throw { status: 400, message: 'Result (HomeWin, AwayWin, Draw) is required to finalize.' };
        }
        const finalizationResult = await bettingApiService.finalizeEvent(eventId, result);
        // Use message from result or default
        const successMessage = finalizationResult?.message || `Event ${eventId} finalized with result ${result}.`;
        const message = encodeURIComponent(successMessage);
        res.redirect(`/betting?message=${message}&type=success`);
    } catch (err) {
        console.error(`Error finalizing event ${eventId}:`, err);
        const message = encodeURIComponent(`Failed to finalize event ${eventId}: ${err.message}`);
         res.redirect(`/betting?message=${message}&type=error`);
        // Alternatively, render an error page:
        // renderError(res, err, `Failed to finalize event ${eventId}.`);
    }
});


module.exports = router;