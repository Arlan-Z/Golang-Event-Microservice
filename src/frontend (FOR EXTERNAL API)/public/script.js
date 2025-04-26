const baseURL = '/api'; 

async function loadEvents() {
  try {
    const response = await fetch(`${baseURL}/api/events`);
    if (!response.ok) {
      throw new Error('Ошибка загрузки событий');
    }
    const events = await response.json();
    displayEvents(events);
  } catch (error) {
    console.error('Ошибка загрузки событий:', error);
  }
}

function displayEvents(events) {
  const list = document.getElementById('events-list');
  list.innerHTML = '';
  
  events.forEach(event => {
    const li = document.createElement('li');
    li.textContent = event.title || 'Без названия';
    list.appendChild(li);
  });
}

document.addEventListener('DOMContentLoaded', loadEvents);
