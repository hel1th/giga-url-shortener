import http from 'k6/http';
import { check } from 'k6';

export const options = {
  scenarios: {
  ceiling: {
    executor: 'ramping-arrival-rate',
    startRate: 1000,
    timeUnit: '1s',
    preAllocatedVUs: 200,
    maxVUs: 1000,
    stages: [
      { target: 2000, duration: '30s' },
      { target: 4000, duration: '30s' },
      { target: 8000, duration: '30s' },
      { target: 12000, duration: '30s' },
    ],
  },
  },
  thresholds: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(95)<50'], // подставь свою цель, мс
  },
};

const BASE = __ENV.BASE_URL || 'http://localhost:8080';

export function setup() {
  const codes = [];
  for (let i = 0; i < 1000; i++) {
    const res = http.post(
      `${BASE}/api/v1/links`,
      JSON.stringify({ original_url: `https://go.dev/?i=${i}` }),
      { headers: { 'Content-Type': 'application/json' } }
    );
    codes.push(res.json('short_code'));
  }
  return { codes };
}

export default function (data) {
  const code = data.codes[Math.floor(Math.random() * data.codes.length)];
  const res = http.get(`${BASE}/r/${code}`, { redirects: 0 });
  check(res, { 'status 302': (r) => r.status === 302 });
}