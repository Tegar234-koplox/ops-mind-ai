import {defineConfig} from '@playwright/test'
import fs from 'node:fs'
import path from 'node:path'
const root=path.resolve(import.meta.dirname,'..')
const envFile=path.join(root,'.env')
if(fs.existsSync(envFile))for(const line of fs.readFileSync(envFile,'utf8').split(/\r?\n/)){const m=/^([A-Z_]+)=(.*)$/.exec(line);if(m&&!process.env[m[1]])process.env[m[1]]=m[2]}
export default defineConfig({
 testDir:'./tests',workers:1,fullyParallel:false,timeout:60000,
 use:{baseURL:process.env.APP_ORIGIN||'http://localhost:8080',browserName:'chromium',viewport:{width:1440,height:1000},screenshot:'only-on-failure',trace:'retain-on-failure'},
 webServer:{command:process.platform==='win32'?'bin\\opsmind.exe serve':'./bin/opsmind serve',url:(process.env.APP_ORIGIN||'http://localhost:8080')+'/healthz',reuseExistingServer:!process.env.CI,timeout:120000,cwd:root,env:{...process.env,MODEL_MODE:'mock',WEB_DIR:path.join(root,'web/dist'),SEED_IF_EMPTY:'true',COOKIE_SECURE:'false'}},
})
