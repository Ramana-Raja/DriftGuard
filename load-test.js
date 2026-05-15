import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    scenarios: {
        pure_db_stress: {
            executor: 'ramping-vus',
            startVUs: 0,
            stages: [
                { duration: '30s', target: 150 },
                { duration: '2m', target: 400 },
                { duration: '30s', target: 0 },
            ],
            gracefulRampDown: '15s',
        },
    },
    thresholds: {
        http_req_failed: ['rate<0.01'],
        http_req_duration: ['p(95)<200'],
    },
};

const BASE_URL = '';

export default function () {
    const uniqueId = `db_${__VU}_${Math.floor(Math.random() * 100000)}`;
    const email = `${uniqueId}@driftguard.com`;
    const password = "BrutalPassword123!";
    const jsonParams = { headers: { 'Content-Type': 'application/json' } };

    http.post(`${BASE_URL}/api/register`, JSON.stringify({ email, password }), jsonParams);

    let loginRes = http.post(`${BASE_URL}/api/login`, JSON.stringify({ email, password }), jsonParams);
    if (!check(loginRes, { 'Login authenticated': (r) => r.status === 200 }) || !loginRes.json('token')) {
        sleep(0.1);
        return;
    }

    const token = loginRes.json('token');
    const authParams = {
        headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json',
        },
    };

    let postProj = http.post(`${BASE_URL}/api/projects`, JSON.stringify({
        name: `Project_${uniqueId}`,
        description: "Phase 1 Isolated DB Test"
    }), authParams);

    check(postProj, { 'Project row committed instantly': (r) => r.status === 200 });

    sleep(0.1);
}