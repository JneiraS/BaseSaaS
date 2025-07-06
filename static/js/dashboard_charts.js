document.addEventListener('DOMContentLoaded', function() {
    console.log("DEBUG: dashboard_charts.js loaded.");
    console.log("DEBUG: typeof Chart:", typeof Chart);
    console.log("DEBUG: Chart object:", Chart);

    // Fonction générique pour récupérer les données et créer un graphique
    async function fetchDataAndCreateChart(url, chartId, chartType, labels, dataKey, title) {
        try {
            const response = await fetch(url, { credentials: 'include' });
            const data = await response.json();

            let chartData = [];
            let chartLabels = [];

            if (chartId === 'membersChart') {
                // Pour les membres par statut
                chartLabels = Object.keys(data.members_by_status);
                chartData = Object.values(data.members_by_status);
            } else if (chartId === 'financeChart') {
                chartLabels = ['Revenus', 'Dépenses', 'Solde Net'];
                chartData = [data.total_income, data.total_expenses, data.net_balance];
            } else if (chartId === 'eventsChart') {
                chartLabels = ['Total Événements'];
                chartData = [data.total_events];
            } else if (chartId === 'documentsChart') {

            new Chart(document.getElementById(chartId), {
                type: chartType,
                data: {
                    labels: chartLabels,
                    datasets: [{
                        label: title,
                        data: chartData,
                        backgroundColor: [
                            'rgba(255, 99, 132, 0.6)',
                            'rgba(54, 162, 235, 0.6)',
                            'rgba(255, 206, 86, 0.6)',
                            'rgba(75, 192, 192, 0.6)',
                            'rgba(153, 102, 255, 0.6)',
                            'rgba(255, 159, 64, 0.6)'
                        ],
                        borderColor: [
                            'rgba(255, 99, 132, 1)',
                            'rgba(54, 162, 235, 1)',
                            'rgba(255, 206, 86, 1)',
                            'rgba(75, 192, 192, 1)',
                            'rgba(153, 102, 255, 1)',
                            'rgba(255, 159, 64, 1)'
                        ],
                        borderWidth: 1
                    }]
                },
                options: {
                    responsive: true,
                    plugins: {
                        legend: {
                            position: 'top',
                        },
                        title: {
                            display: true,
                            text: title
                        }
                    }
                }
            });
        } catch (error) {
            console.error(`Erreur lors de la récupération des données pour ${chartId}:`, error.name, error.message, error);
        }
    }

    // Appels pour chaque graphique
    fetchDataAndCreateChart('/api/stats/members', 'membersChart', 'pie', [], '', 'Statistiques des Membres');
    fetchDataAndCreateChart('/api/stats/finance', 'financeChart', 'bar', [], '', 'Statistiques Financières');
    fetchDataAndCreateChart('/api/stats/events', 'eventsChart', 'bar', [], '', 'Statistiques des Événements');
    
    fetchDataAndCreateChart('/api/stats/documents', 'documentsChart', 'bar', [], '', 'Statistiques des Documents');
});