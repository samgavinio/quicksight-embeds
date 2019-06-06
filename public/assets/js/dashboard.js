import { embedDashboard } from 'amazon-quicksight-embedding-sdk';

class Dashboard {
    constructor(containerId, url) {
        const containerDiv = document.getElementById(containerId);
        const options = {
            url,
            container: containerDiv,
            scrolling: 'yes',
            height: '400px'
        };
        this.dashboard = embedDashboard(options);
        this.dashboard.on('error', this.onError);
        this.dashboard.on('load', this.onDashboardLoad);
    }

    updateParameter(parameters) {
        this.dashboard.setParameters(parameters);
    }

    onError(err) {
        console.log(err);
    }

    onDashboardLoad() {
        console.log(`Dashboard succesfully loaded.`);
    }
}

export default Dashboard;
