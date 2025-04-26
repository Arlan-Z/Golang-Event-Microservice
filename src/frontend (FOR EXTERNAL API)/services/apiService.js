const axios = require('axios');
require('dotenv').config(); // Load .env variables

const apiClient = axios.create({
    baseURL: process.env.API_BASE_URL,
    timeout: 10000, // 10 second timeout
    headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json'
    }
});

// Helper to handle API errors
const handleApiError = (error, functionName) => {
    console.error(`API Error in ${functionName}:`, error.response?.data || error.message);
    // Rethrow a more specific error or return a structured error object
    const message = error.response?.data?.message || error.response?.data || error.message || 'An unknown API error occurred';
    const status = error.response?.status || 500;
    const err = new Error(message);
    err.status = status;
    throw err; // Re-throw the error to be caught by route handlers
};

// --- Event Endpoints ---

const getAllEvents = async () => {
    try {
        const response = await apiClient.get('/Events/all');
        return response.data;
    } catch (error) {
        handleApiError(error, 'getAllEvents');
    }
};

const getEventById = async (id) => {
    try {
        const response = await apiClient.get(`/Events/${id}`);
        return response.data;
    } catch (error) {
        handleApiError(error, 'getEventById');
    }
};

const createEvent = async (eventData) => {
    // Ensure required fields have sensible defaults if not provided
    const typeAsInt = eventData.type !== undefined ? parseInt(eventData.type, 10) : 0;
    const payload = {
        eventName: eventData.eventName || 'Unnamed Event',
        type: typeAsInt, // Assuming 0 maps to 'Other' or a default Enum value
        homeTeam: eventData.homeTeam || 'Home Team',
        awayTeam: eventData.awayTeam || 'Away Team',
        eventStartDate: eventData.eventStartDate || new Date().toISOString(),
        eventEndDate: eventData.eventEndDate || new Date().toISOString(),
        // eventSubscribers: [] // Subscribers usually added later
    };
    try {
        const response = await apiClient.post('/Events/create', payload);
        return response.data;
    } catch (error) {
        handleApiError(error, 'createEvent');
    }
};

const endEvent = async (id) => {
    try {
        const response = await apiClient.post(`/Events/${id}/end`);
        return response.data;
    } catch (error) {
        handleApiError(error, 'endEvent');
    }
};

const cancelEvent = async (id) => {
    try {
        const response = await apiClient.post(`/Events/${id}/cancel`);
        return response.data; // Often just a success message
    } catch (error) {
        handleApiError(error, 'cancelEvent');
    }
};

const subscribeToEvent = async (id, callbackUrl) => {
    try {
        const response = await apiClient.post(`/Events/${id}/subscribe`, { callbackUrl });
        return response.data; // Success message
    } catch (error) {
        handleApiError(error, 'subscribeToEvent');
    }
};

const deleteEvent = async (id) => {
    try {
        // DELETE returns 204 No Content on success, data might be empty
        await apiClient.delete(`/Events/${id}/delete`);
        return { message: `Event ${id} deleted successfully.` };
    } catch (error) {
        handleApiError(error, 'deleteEvent');
    }
};

// --- Event Detail Endpoints ---

const getEventDetails = async (id) => {
    try {
        const response = await apiClient.get(`/Events/${id}/details`);
        return response.data;
    } catch (error) {
        // Handle 404 specifically if needed, otherwise general handler works
        if (error.response?.status === 404) {
             const err = new Error(`Event details not found for ID ${id}`);
             err.status = 404;
             throw err;
        }
        handleApiError(error, 'getEventDetails');
    }
};

const createEventDetail = async (id, detailData) => {
     // Provide defaults for optional fields expected by the API
     const payload = {
        roundNumber: detailData.roundNumber ? parseInt(detailData.roundNumber, 10) : 1,
        homeTeamScore: detailData.homeTeamScore ? parseInt(detailData.homeTeamScore, 10) : 0,
        awayTeamScore: detailData.awayTeamScore ? parseInt(detailData.awayTeamScore, 10) : 0,
     };
    try {
        const response = await apiClient.post(`/Events/${id}/details/create`, payload);
        return response.data;
    } catch (error) {
        handleApiError(error, 'createEventDetail');
    }
};

const updateEventDetail = async (id, detailData) => {
    // Prepare payload - API likely expects roundNumber to identify which detail to update
    const payload = {
        roundNumber: detailData.roundNumber ? parseInt(detailData.roundNumber, 10) : undefined, // Required
        homeTeamScore: detailData.homeTeamScore !== undefined ? parseInt(detailData.homeTeamScore, 10) : undefined, // Optional
        awayTeamScore: detailData.awayTeamScore !== undefined ? parseInt(detailData.awayTeamScore, 10) : undefined, // Optional
    };
    // Remove undefined properties as the API might interpret them as explicit nulls
    Object.keys(payload).forEach(key => payload[key] === undefined && delete payload[key]);

    if (!payload.roundNumber) {
        throw new Error("Round number is required to update event details.");
    }

    try {
        // PATCH is the correct verb according to the C# controller
        const response = await apiClient.patch(`/Events/${id}/details`, payload);
        return response.data;
    } catch (error) {
        handleApiError(error, 'updateEventDetail');
    }
};


module.exports = {
    getAllEvents,
    getEventById, // Kept for completeness, though details endpoint might be more useful
    createEvent,
    endEvent,
    cancelEvent,
    subscribeToEvent,
    deleteEvent,
    getEventDetails,
    createEventDetail,
    updateEventDetail // Note: The C# controller has PATCH, we'll call PATCH here.
};