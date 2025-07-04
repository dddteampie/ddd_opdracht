import React, { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { List, Card, Spin, message, Button } from 'antd';
import Container from '../common/Container';
import {
  startProductAanvraag,
  getPassendeProductenLijst,
  kiesProduct,
} from '../services/behoefteService'; // Pas aan indien nodig
const AdviesPage = () => {
  const { clientId, aanvraagId } = useParams();
  const [producten, setProducten] = useState([]);
  const [loading, setLoading] = useState(false);
  const [aanvraagDetails, setAanvraagDetails] = useState(null); // Om de gekozen categorie en behoeftebeschrijving te weten

  useEffect(() => {
    if (clientId && aanvraagId) {
      // Normaal zou je hier eerst de aanvraag details ophalen, inclusief de gekozen categorie
      // Voor nu simuleren we dit.
      fetchAanvraagDetails(clientId, aanvraagId);
    }
  }, [clientId, aanvraagId]);

  const fetchAanvraagDetails = async (clientId, aanvraagId) => {
    // Deze functie zou een API call moeten doen om de volledige aanvraag op te halen
    // zodat we de behoeftebeschrijving en gekozen categorie ID weten.
    // Voorbeeld: const details = await getAanvraagById(aanvraagId);
    // Voorlopig hardcoded, dit moet je later vervangen.
    const mockAanvraagDetails = {
      clientid: clientId,
      behoefteid: aanvraagId, // In dit voorbeeld gebruiken we behoefteId als aanvraagId
      behoeftebeschrijving: "Behoefte om langer thuis te blijven wonen", // Placeholder
      gekozencategorieid: 1 // Placeholder: ID van de gekozen categorie
    };
    setAanvraagDetails(mockAanvraagDetails);
    
    // Zodra we de details hebben, starten we de product aanvraag
    handleStartProductAanvraag(mockAanvraagDetails);
  };


  const handleStartProductAanvraag = async (details) => {
    setLoading(true);
    try {
      const inputData = {
        client_id: details.clientid,
        behoefte_beschrijving: details.behoeftebeschrijving,
        gekozen_categorie_id: details.gekozencategorieid,
      };
      await startProductAanvraag(inputData);
      message.success("Productaanvraag gestart. Producten ophalen...");
      
      const fetchedProducts = await getPassendeProductenLijst(details.clientid);
      setProducten(fetchedProducts.productlijst);
      message.success("Producten succesvol opgehaald.");

    } catch (error) {
      message.error("Fout bij opvragen productaanbeveling: " + error.message);
      console.error("Error starting product request:", error);
    } finally {
      setLoading(false);
    }
  };

  const handleKiesProduct = async (productEAN) => {
    if (!aanvraagDetails) {
      message.error("Aanvraagdetails niet geladen.");
      return;
    }
    setLoading(true);
    try {
      const inputData = {
        client_id: aanvraagDetails.clientid,
        behoefte_id: aanvraagDetails.behoefteid,
        product_ean: productEAN
      };
      await kiesProduct(inputData);
      message.success(`Product ${productEAN} gekozen.`);
      // Hier kun je verdere acties ondernemen, bijv. naar een bevestigingspagina
    } catch (error) {
      message.error("Fout bij het kiezen van product: " + error.message);
      console.error("Error choosing product:", error);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <Container>
        <Spin size="large" tip="Producten laden..." />
      </Container>
    );
  }

  return (
    <Container>
      <h1>Product Aanbevelingen voor Client: {clientId}</h1>
      {producten.length > 0 ? (
        <List
          grid={{ gutter: 16, xs: 1, sm: 2, md: 3, lg: 4, xl: 4, xxl: 4 }}
          dataSource={producten}
          renderItem={(product) => (
            <List.Item>
              <Card
                title={product.Naam}
                actions={[
                  <Button type="primary" onClick={() => handleKiesProduct(product.EAN)}>
                    Kies Product
                  </Button>,
                ]}
              >
                <p>EAN: {product.EAN}</p>
                <p>Prijs: €{product.Prijs ? product.Prijs.toFixed(2) : 'N/A'}</p>
                <p>Beschrijving: {product.Beschrijving || 'Geen beschrijving'}</p>
                {/* Voeg hier meer productdetails toe */}
              </Card>
            </List.Item>
          )}
        />
      ) : (
        <p>Geen product aanbevelingen gevonden voor deze aanvraag.</p>
      )}
    </Container>
  );
};

export default AdviesPage;