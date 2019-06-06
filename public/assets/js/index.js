import Dashboard from './dashboard';
import '../css/sb-admin-2.css';

let dashboard;

window.document.addEventListener('DOMContentLoaded', () => {
    dashboard = new Dashboard('dashboard-container', document.getElementById('dashboard-url').value);
});

document.getElementById('change-parameter-btn').addEventListener('click', () => {
    console.log(document.getElementById('parameter-value').value);
    dashboard.updateParameter({
        InstanceID: document.getElementById('parameter-value').value
    });
});
