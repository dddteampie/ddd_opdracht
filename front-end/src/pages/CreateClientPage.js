import React, { useState } from 'react';
import { Form, Input, Button, DatePicker, message, Card, Spin, Typography } from 'antd';
import { useHistory } from 'react-router-dom'; 
import moment from 'moment'; 
import Container from '../common/Container';
import { addClient,  } from '../services/clientService'; 
import { createOnderzoek, createZorgdossier } from '../services/behoefteService';

const { Title } = Typography;

const CreateClientPage = () => {
    const [form] = Form.useForm();
    const [loading, setLoading] = useState(false);
    const history = useHistory();

    const handleCreateClient = async (values) => {
        setLoading(true);
        try {
            const clientData = {
                naam: values.naam,
                adres: values.adres,
                geboortedatum: values.geboortedatum.toISOString(),
            };
            const newClient = await addClient(clientData);
            message.success(`Client ${newClient.naam} succesvol aangemaakt!`);

            const zorgdossierData = {
                client_id: newClient.id,
                situatie: "Initiële situatie na aanmelding client", 
            };
            const newZorgdossier = await createZorgdossier(zorgdossierData);
            message.success(`Zorgdossier voor ${newClient.naam} succesvol aangemaakt!`);

            const vandaag = moment();
            const onderzoekData = {
                zorgdossier_id: newZorgdossier.id, 
                begin_datum: vandaag.toISOString(),
                eind_datum: vandaag.add(1, 'year').toISOString(), 
            };
            await createOnderzoek(onderzoekData);
            message.success(`Onderzoek voor ${newClient.naam} succesvol aangemaakt!`);

            history.push(`/clients/${newClient.id}`);
            form.resetFields(); 
        } catch (error) {
            message.error("Fout bij het aanmaken van client, zorgdossier of onderzoek.");
            console.error("Error creating client, zorgdossier, or onderzoek:", error);
        } finally {
            setLoading(false);
        }
    };

    return (
        <Container>
            <Card>
                <Title level={2} style={{ textAlign: 'center' }}>Nieuwe Cliënt Aanmaken</Title>
                <Spin spinning={loading} tip="Client, zorgdossier en onderzoek aanmaken...">
                    <Form
                        form={form}
                        layout="vertical"
                        onFinish={handleCreateClient}
                    >
                        <Form.Item
                            label="Naam Cliënt"
                            name="naam"
                            rules={[{ required: true, message: 'Vul de naam van de cliënt in!' }]}
                        >
                            <Input placeholder="Volledige naam" />
                        </Form.Item>
                        <Form.Item
                            label="Adres"
                            name="adres"
                            rules={[{ required: true, message: 'Vul het adres in!' }]}
                        >
                            <Input placeholder="Straatnaam, huisnummer, postcode, plaats" />
                        </Form.Item>
                        <Form.Item
                            label="Geboortedatum"
                            name="geboortedatum"
                            rules={[{ required: true, message: 'Selecteer de geboortedatum!' }]}
                        >
                            <DatePicker style={{ width: '100%' }} format="YYYY-MM-DD" />
                        </Form.Item>
                        <Form.Item>
                            <Button type="primary" htmlType="submit" loading={loading} block>
                                Cliënt Aanmaken en Initialiseren
                            </Button>
                        </Form.Item>
                    </Form>
                </Spin>
            </Card>
        </Container>
    );
};

export default CreateClientPage;