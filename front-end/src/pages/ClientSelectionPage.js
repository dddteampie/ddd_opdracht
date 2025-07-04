// src/pages/ClientSelectionPage.js (Gecorrigeerde versie)
import React, { useState, useEffect } from 'react';
import { useHistory } from 'react-router-dom'; // Voor React Router v5
import { Button, Input, List, Card, Spin, message } from 'antd';
import Container from '../common/Container';
import { getClients, addClient } from '../services/clientService'; // Zorg dat deze paden correct zijn
import moment from 'moment'; // Voor het formatteren van de geboortedatum indien nodig

const ClientSelectionPage = () => {
  const history = useHistory();
  const [clients, setClients] = useState([]);
  const [newClientName, setNewClientName] = useState('');
  const [loading, setLoading] = useState(false);
  // activeClient state is hier niet direct nodig voor de navigatielogica,
  // maar kan nuttig zijn voor andere UI-doeleinden, dus ik laat hem staan.
  const [activeClient, setActiveClient] = useState(null);

  useEffect(() => {
    fetchClients();
  }, []);

  const fetchClients = async () => {
    setLoading(true);
    try {
      const fetchedClients = await getClients();
      setClients(fetchedClients || []); // Zorg ervoor dat clients altijd een array is
    } catch (error) {
      message.error("Fout bij het laden van cliënten.");
      console.error("Error fetching clients:", error);
    } finally {
      setLoading(false);
    }
  };

  const handleAddClient = async () => {
    if (!newClientName.trim()) {
      message.warning("Voer een naam in voor de nieuwe cliënt.");
      return;
    }
    setLoading(true);
    try {
      const newClientData = {
        naam: newClientName,
        adres: "Onbekend", // Tijdelijke placeholder
        geboortedatum: moment().toISOString(), // Gebruik moment voor consistente ISO-format
      };
      // Je backend reageert met `{ "clientId": "..." }`
      const response = await addClient(newClientData);

      if (response && response.clientId) {
        const addedClient = {
          // Gebruik response.clientId zoals je Go backend die levert
          id: response.clientId,
          naam: newClientName,
          adres: newClientData.adres, // Gebruik de data die je hebt verstuurd
          geboortedatum: newClientData.geboortedatum,
        };
        setClients([...clients, addedClient]);
        setNewClientName('');
        message.success(`Cliënt "${addedClient.naam}" succesvol toegevoegd.`);

        // **BELANGRIJK: Navigeer nu naar de ClientDetailPage voor initialisatie**
        history.push(`/clients/${addedClient.id}`);
      } else {
        message.error("Fout bij het toevoegen van cliënt: Geen geldig ID ontvangen.");
      }

    } catch (error) {
      message.error("Fout bij het toevoegen van cliënt: " + error.message);
      console.error("Error adding client:", error);
    } finally {
      setLoading(false);
    }
  };

  const handleSelectClient = (client) => {
    setActiveClient(client); // Optioneel, voor UI weergave
    
    message.success(`Cliënt "${client.naam}" geselecteerd.`);
    // **Navigeer naar de ClientDetailPage als tussenstap**
    history.push(`/clients/${client.id}`);
  };

  if (loading) {
    return (
      <Container>
        <Spin size="large" tip="Cliënten laden..." />
      </Container>
    );
  }

  return (
    <Container>
      <h1>Cliënt Selectie</h1>
      <Card title="Nieuwe Cliënt Toevoegen" style={{ marginBottom: '20px' }}>
        <Input
          placeholder="Naam van de nieuwe cliënt"
          value={newClientName}
          onChange={(e) => setNewClientName(e.target.value)}
          style={{ width: 'calc(100% - 100px)', marginRight: '10px' }}
        />
        <Button type="primary" onClick={handleAddClient}>
          Toevoegen
        </Button>
      </Card>

      <Card title="Bestaande Cliënten">
        <List
          bordered
          dataSource={clients}
          renderItem={(client) => (
            <List.Item
              actions={[
                <Button key="select" onClick={() => handleSelectClient(client)}>
                  Selecteer
                </Button>,
              ]}
            >
              <List.Item.Meta
                title={client.naam}
                description={`ID: ${client.id}`}
              />
            </List.Item>
          )}
        />
      </Card>
    </Container>
  );
};

export default ClientSelectionPage;