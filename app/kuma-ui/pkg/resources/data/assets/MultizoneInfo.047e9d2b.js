import{P as m,N as u,b as l,ck as _,cn as d,E as g,i as n,o as K,c as f,w as e,a,e as t,l as o,t as k}from"./index.60b0f0ac.js";const y={name:"MultizoneInfo",components:{KButton:m,KEmptyState:u,KIcon:l},data(){return{productName:_}},computed:{...d({kumaDocsVersion:"config/getKumaDocsVersion"})}},N=o("p",null,[t(`
        To access this page, you must be running in `),o("strong",null,"Multi-Zone"),t(` mode.
      `)],-1);function B(s,M,V,b,c,h){const r=n("KIcon"),i=n("KButton"),p=n("KEmptyState");return K(),f(p,null,{title:e(()=>[a(r,{class:"mb-3",icon:"dangerCircleOutline",size:"42"}),t(),o("p",null,k(c.productName)+" is running in Standalone mode.",1)]),message:e(()=>[N]),cta:e(()=>[a(i,{to:`https://kuma.io/docs/${s.kumaDocsVersion}/documentation/deployments/`,target:"_blank",appearance:"primary"},{default:e(()=>[t(`
        Learn More
      `)]),_:1},8,["to"])]),_:1})}const E=g(y,[["render",B]]);export{E as M};
