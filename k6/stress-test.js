import { check, sleep } from 'k6';
import http from 'k6/http';
import faker from 'k6/x/faker';

export const options = {
  stages: [
    { duration: '30s', target: 50 },   // ramp-up inicial
    { duration: '30s', target: 200 },  // comeca o stress
    { duration: '60s', target: 300 },  // pico
    { duration: '30s', target: 300 },  // mantem
    { duration: '30s', target: 0 },    // ramp-down
  ]
};


export default function() {
  const payload = JSON.stringify({
    "flight": faker.strings.digitN(8),
    "day": faker.time.dateRange("2026-01-01", "2026-12-31", "yyyy-MM-dd"),
    "user": faker.person.firstName(),
    "ft": false
  });

  // const payload = JSON.stringify({
  //   "flight": "11111111",
  //   "day": "2026-01-01",
  //   "user": faker.person.firstName(),
  //   "ft": false
  // });

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  const response = http.post('http://127.0.0.1:8080/buyTicket', payload, params)
  check(response, { 'status is 200': (r) => r.status === 200 || r.status === 504 });
  sleep(0.2);

}
