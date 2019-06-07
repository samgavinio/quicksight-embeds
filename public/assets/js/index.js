import Dashboard from './dashboard';
import '../css/sb-admin-2.css';

let dashboard;

window.document.addEventListener('DOMContentLoaded', () => {
    dashboard = new Dashboard('dashboard-container', document.getElementById('dashboard-url').value);
});

document.getElementById('change-parameter-btn').addEventListener('click', () => {
    // Modify the parameter argument according to what your dashboard expects
    dashboard.updateParameter({
        InstanceID: document.getElementById('parameter-value').value
    });
});
