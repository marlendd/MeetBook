import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '30s', target: 50 },   
    { duration: '1m',  target: 100 },  
    { duration: '30s', target: 0 },    
  ],
  thresholds: {
    'http_req_failed':                    ['rate<0.001'],
    'http_req_duration{endpoint:slots}':  ['p(95)<200'],
  },
};

const BASE = 'http://localhost:8080';

export function setup() {
  const adminToken = http.post(`${BASE}/dummyLogin`,
    JSON.stringify({ role: 'admin' }),
    { headers: { 'Content-Type': 'application/json' } }
  ).json('token');

  const userToken = http.post(`${BASE}/dummyLogin`,
    JSON.stringify({ role: 'user' }),
    { headers: { 'Content-Type': 'application/json' } }
  ).json('token');

  const adminHeaders = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${adminToken}`,
  };

  const room = http.post(`${BASE}/rooms/create`,
    JSON.stringify({ name: 'Load Test Room' }),
    { headers: adminHeaders }
  ).json('room');

  http.post(`${BASE}/rooms/${room.id}/schedule/create`,
    JSON.stringify({ daysOfWeek: [1,2,3,4,5,6,7], startTime: '09:00', endTime: '18:00' }),
    { headers: adminHeaders }
  );

  return { roomId: room.id, userToken, adminToken };
}

export default function (data) {
  const userHeaders = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${data.userToken}`,
  };

  const today = new Date().toISOString().split('T')[0];

  // Самый нагруженный эндпоинт — список слотов
  const slotsRes = http.get(
    `${BASE}/rooms/${data.roomId}/slots/list?date=${today}`,
    { headers: userHeaders, tags: { endpoint: 'slots' } }
  );

  check(slotsRes, {
    'slots 200': (r) => r.status === 200,
  });

  sleep(0.1);
}
