const axios = require('axios');
require('dotenv').config();

const bettingApiClient = axios.create({
    baseURL: process.env.BETTING_API_BASE_URL, 
    timeout: 15000, // Slightly longer timeout might be needed
    headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json'
    }
});

// Reusable error handler (similar to the previous one)
const handleApiError = (error, functionName) => {
    console.error(`Betting API Error in ${functionName}:`, error.response?.data || error.message);
    const message = error.response?.data?.title || error.response?.data?.message || error.response?.data || error.message || 'An unknown betting API error occurred';
    const status = error.response?.status || 500;
    const err = new Error(message);
    err.status = status;
    // Include original error details if available (useful for debugging)
    if (error.response?.data) {
        err.details = error.response.data;
    }
    throw err;
};

// --- Betting API Functions ---

const getActiveEvents = async () => {
    try {
        console.log(`Fetching active events from: ${process.env.BETTING_API_BASE_URL}/events`);
        const response = await bettingApiClient.get('/events');
        return response.data || []; // Return empty array if data is null/undefined
    } catch (error) {
        // Handle cases where the API might be down entirely
        if (error.code === 'ECONNREFUSED') {
             throw new Error(`Connection refused when trying to reach Betting API at ${process.env.BETTING_API_BASE_URL}. Is it running?`);
        }
        handleApiError(error, 'getActiveEvents');
    }
};

const placeBet = async (betData) => {
    // Validate required fields before sending
    if (!betData.userId || !betData.eventId || !betData.amount || !betData.predictedOutcome) {
        const missing = [
            !betData.userId && "userId",
            !betData.eventId && "eventId",
            !betData.amount && "amount",
            !betData.predictedOutcome && "predictedOutcome"
        ].filter(Boolean).join(', ');
        throw new Error(`Missing required bet data: ${missing}`);
    }
    if (parseFloat(betData.amount) <= 0) {
         throw new Error('Bet amount must be greater than 0.');
    }
     // Basic UUID check (simple regex, not foolproof)
     const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
     if (!uuidRegex.test(betData.userId)) {
         throw new Error('Invalid format for userId (must be UUID).');
     }
     if (!uuidRegex.test(betData.eventId)) {
         throw new Error('Invalid format for eventId (must be UUID).');
     }

    const payload = {
        userId: betData.userId,
        eventId: betData.eventId,
        amount: parseFloat(betData.amount), // Ensure amount is a number
        predictedOutcome: betData.predictedOutcome // Should be "HomeWin", "AwayWin", or "Draw"
    };
    console.log('Placing bet with payload:', payload);
    try {
        const response = await bettingApiClient.post('/bets', payload);
        return response.data; // Returns the created BetDTO
    } catch (error) {
        handleApiError(error, 'placeBet');
    }
};

const finalizeEvent = async (eventId, result) => {
     if (!eventId || !result) {
         throw new Error("Both eventId and result are required to finalize.");
     }
     const validResults = ["HomeWin", "AwayWin", "Draw"];
     if (!validResults.includes(result)) {
         throw new Error(`Invalid result value: ${result}. Must be one of ${validResults.join(', ')}`);
     }
      // Basic UUID check
     const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
      if (!uuidRegex.test(eventId)) {
         throw new Error('Invalid format for eventId (must be UUID).');
     }

    const payload = { result };
    console.log(`Finalizing event ${eventId} with result ${result}`);
    try {
        // Note: API expects POST request, response is 200 OK on success
        const response = await bettingApiClient.post(`/events/${eventId}/finalize`, payload);
        // API returns 200 OK, response.data might be empty or contain a confirmation
        return response.data || { message: `Event ${eventId} finalization initiated with result ${result}.` };
    } catch (error) {
        handleApiError(error, 'finalizeEvent');
    }
};

// Optional: Add functions for health/readiness if needed by UI
const checkHealth = async () => {
    try {
        await bettingApiClient.get('/healthz');
        return { status: 'OK' };
    } catch (error) {
         console.error("Health check failed:", error.message);
        return { status: 'Error', message: error.message };
    }
}

const checkReadiness = async () => {
     try {
        await bettingApiClient.get('/readyz');
        return { status: 'Ready' };
    } catch (error) {
         console.error("Readiness check failed:", error.message);
         const status = error.response?.status === 503 ? 'Unavailable' : 'Error';
        return { status: status, message: error.message };
    }
}


module.exports = {
    getActiveEvents,
    placeBet,
    finalizeEvent,
    checkHealth,
    checkReadiness
};