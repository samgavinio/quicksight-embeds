import QuickSightEmbedding from 'amazon-quicksight-embedding-sdk';

const onError = (err) => console.log(err);
const onDashboardLoad= (x) => console.log(x);

const EmbedDashboard = () => {
    const containerDiv = document.getElementById('dashboard-container');
    const options = {
        url: document.getElementById('dashboard-url"').val,
        container: containerDiv,
        parameters: {
            country: 'United States'
        },
        scrolling: 'no',
        height: '700px',
        width: '1000px'
    };
    const dashboard = QuickSightEmbedding.embedDashboard(options);
    dashboard.on('error', onError);
    dashboard.on('load', onDashboardLoad);
};

export default EmbedDashboard;
