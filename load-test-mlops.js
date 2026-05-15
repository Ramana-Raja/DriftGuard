import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    scenarios: {
        focused_mlops_ingestion: {
            executor: 'constant-vus',
            vus: 20,
            duration: '2m',
        },
    },
    thresholds: {
        http_req_failed: ['rate<0.05'],
        http_req_duration: ['p(95)<800'],
    },
};

const BASE_URL = '';
const DATASET_URL = 'https://raw.githubusercontent.com/Ramana-Raja/testing_upload/refs/heads/main/iris_dataset.csv';
const dummyModelFile = open('./v1.pkl', 'b');

export default function () {
    const uniqueId = `mlops_${__VU}_${Math.floor(Math.random() * 100000)}`;
    const email = `${uniqueId}@driftguard.com`;
    const password = "BrutalPipelinePassword123!";
    const jsonParams = { headers: { 'Content-Type': 'application/json' } };

    http.post(`${BASE_URL}/api/register`, JSON.stringify({ email, password }), jsonParams);
    let loginRes = http.post(`${BASE_URL}/api/login`, JSON.stringify({ email, password }), jsonParams);
    if (!loginRes.json('token')) return;

    const token = loginRes.json('token');
    const authHeaders = { 'Authorization': `Bearer ${token}` };

    let postProj = http.post(`${BASE_URL}/api/projects`, JSON.stringify({ name: `Project_${uniqueId}` }), {
        headers: Object.assign({}, authHeaders, { 'Content-Type': 'application/json' })
    });
    if (postProj.status !== 200 || !postProj.json('id')) return;

    const projectId = postProj.json('id');

    const multipartData = {
        model: http.file(dummyModelFile, 'model.pkl', 'application/octet-stream'),
        dataset_link: DATASET_URL,
    };

    let uploadRes = http.post(`${BASE_URL}/api/projects/${projectId}/model`, multipartData, { headers: authHeaders });
    check(uploadRes, { 'Full MLOps Pipeline Successful': (r) => r.status === 200 });

    sleep(0.5);
}