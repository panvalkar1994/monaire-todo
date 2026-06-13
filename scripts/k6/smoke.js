import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const API = `${BASE_URL}/api/v1/todos`;

export const options = {
    vus: 1,
    iterations: 1,
    thresholds: {
        checks: ['rate==1'],
    },
};

export default function () {
    // Create
    const createRes = http.post(
        API,
        JSON.stringify({
            text: 'k6 smoke test',
            due_date: '2026-06-15',
            completed: false,
        }),
        { headers: { 'Content-Type': 'application/json' } },
    );
    check(createRes, {
        'create status 201': (r) => r.status === 201,
    });
    const todo = createRes.json();
    const id = todo.id;

    // Get
    const getRes = http.get(`${API}/${id}`);
    check(getRes, {
        'get status 200': (r) => r.status === 200,
        'get returns id': (r) => r.json('id') === id,
    });

    // List (default)
    const listRes = http.get(API);
    check(listRes, {
        'list status 200': (r) => r.status === 200,
        'list is array': (r) => Array.isArray(r.json()),
    });

    // List include_completed=true
    const listAllRes = http.get(`${API}?include_completed=true`);
    check(listAllRes, {
        'list all status 200': (r) => r.status === 200,
    });

    // List invalid include_completed → 417
    const invalidListRes = http.get(`${API}?include_completed=maybe`);
    check(invalidListRes, {
        'invalid include_completed status 417': (r) => r.status === 417,
        'invalid include_completed error': (r) =>
            r.json('error') === 'include_completed allowed values: true|false|empty',
    });

    // Patch
    const patchRes = http.patch(
        `${API}/${id}`,
        JSON.stringify({ completed: true }),
        { headers: { 'Content-Type': 'application/json' } },
    );
    check(patchRes, {
        'patch status 200': (r) => r.status === 200,
        'patch completed true': (r) => r.json('completed') === true,
    });

    // PUT replace (same body → X-No-Changes)
    const replaceBody = JSON.stringify({
        text: todo.text,
        due_date: todo.due_date,
        completed: true,
    });
    const replaceRes = http.put(`${API}/${id}`, replaceBody, {
        headers: { 'Content-Type': 'application/json' },
    });
    check(replaceRes, {
        'replace status 200': (r) => r.status === 200,
    });

    // Delete
    const deleteRes = http.del(`${API}/${id}`);
    check(deleteRes, {
        'delete status 204': (r) => r.status === 204,
    });

    // Confirm gone
    const goneRes = http.get(`${API}/${id}`);
    check(goneRes, {
        'get after delete status 404': (r) => r.status === 404,
    });

    sleep(0.1);
}
