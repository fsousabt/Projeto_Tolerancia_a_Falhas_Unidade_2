import { check, sleep } from 'k6';
import http from 'k6/http';
import faker from 'k6/x/faker';

export const options = {
  stages: [
    { duration: '30s', target: 50},
    { duration: '2m', target: 50},
    { duration: '30s', target: 0 },
  ]

}

if (__ENV.ft !== "true" && __ENV.ft !== "false") throw new Error("Variavel ft deve ser true ou false, cheque o README para mais detalhes.");
const ft = __ENV.ft === "true" ? true : false

export default function() {
  const payload = JSON.stringify({
    "flight": faker.strings.digitN(8),
    "day": faker.time.dateRange("2026-01-01", "2026-12-31", "yyyy-MM-dd"),
    "user": faker.person.firstName(),
    "ft": ft
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  const response = http.post('http://127.0.0.1:8080/buyTicket', payload, params)
  check(response, { 'status is 200': (r) => r.status === 200 || r.status === 504 });

  sleep(0.2);

}
