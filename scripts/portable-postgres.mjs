import fs from 'node:fs';import path from 'node:path';import EmbeddedPostgres from '../.tools/pg/node_modules/embedded-postgres/dist/index.js';
const root=path.resolve(import.meta.dirname,'..'),dir=path.join(root,'.local','pgdata');
const pg=new EmbeddedPostgres({databaseDir:dir,user:'opsmind',password:'local-only-database',port:55432,persistent:true,initdbFlags:['--encoding=UTF8','--locale=C'],postgresFlags:['-c','listen_addresses=127.0.0.1'],onLog:()=>{},onError:message=>console.error(String(message))});
if(!fs.existsSync(path.join(dir,'PG_VERSION')))await pg.initialise();await pg.start();
const client=pg.getPgClient();await client.connect();for(const name of ['opsmind','opsmind_test']){const r=await client.query('SELECT 1 FROM pg_database WHERE datname=$1',[name]);if(!r.rowCount)await client.query('CREATE DATABASE '+name)}await client.end();
console.log('PostgreSQL ready on 127.0.0.1:55432; local databases created.');
process.on('SIGINT',async()=>{await pg.stop();process.exit(0)});
process.on('SIGTERM',async()=>{await pg.stop();process.exit(0)});
setInterval(()=>{},60000);
