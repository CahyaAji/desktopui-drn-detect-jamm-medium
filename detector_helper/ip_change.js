const net = require('net');

// The default connection details from the manual
const DEFAULT_IP = '192.168.1.10'; //
const DEFAULT_PORT = 7;            //

// The NEW IP and Port you want to assign
// Change these values to whatever you need!
const NEW_IP = '192.168.17.102';
const NEW_PORT = 7;

// Format the command string as required by the module
// Note: We replace the dots in the IP with commas for the command format
const ipParts = NEW_IP.split('.');
const command = `$,setServer,${ipParts[0]},${ipParts[1]},${ipParts[2]},${ipParts[3]},${NEW_PORT}\n`;

const client = new net.Socket();

console.log(`Connecting to module at ${DEFAULT_IP}:${DEFAULT_PORT}...`);

client.connect(DEFAULT_PORT, DEFAULT_IP, () => {
    console.log('Connected!');
    
    console.log(`Sending command: ${command.trim()}`);
    client.write(command);
});

// Listen for a response from the module
client.on('data', (data) => {
    const response = data.toString().trim();
    console.log(`Received from module: ${response}`);
    
    if (response === '#,setServer,failed') {
        console.log('Error: Failed to set the new IP address.');
    }
});

// The documentation states that upon success, the connection will drop.
client.on('close', () => {
    console.log('Connection closed.');
    console.log('If the command was successful, the module disconnected and is now on the new IP.');
});

client.on('error', (err) => {
    console.error(`Connection Error: ${err.message}`);
});