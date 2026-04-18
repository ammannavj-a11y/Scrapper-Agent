import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
  stages: [
    { duration: '30s', target: 50 },
    { duration: '60s', target: 100 },
    { duration: '30s', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(99)<2000'],
    http_req_failed: ['rate<0.01'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export default function () {
  const guvid = `GUV-IN-2025-X7K2M9PQ`;
  const res = http.get(`${BASE_URL}/api/v1/guvid/verify?guvid=${guvid}`);
  check(res, {
    'status is 200': (r) => r.status === 200,
    'has trustLevel': (r) => r.json('trustLevel') !== undefined,
  });
  sleep(0.1);
}
