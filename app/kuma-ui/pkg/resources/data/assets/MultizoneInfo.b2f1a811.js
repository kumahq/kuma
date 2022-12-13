import{A as m,N as u,g as l,ck as _,cn as d,E as g,i as n,o as K,c as f,w as t,a,b as e,l as o,t as k}from"./index.0cb244cf.js";const y={name:"MultizoneInfo",components:{KButton:m,KEmptyState:u,KIcon:l},data(){return{productName:_}},computed:{...d({kumaDocsVersion:"config/getKumaDocsVersion"})}},N=o("p",null,[e(`
        To access this page, you must be running in `),o("strong",null,"Multi-Zone"),e(` mode.
      `)],-1);function B(s,M,V,h,c,D){const r=n("KIcon"),i=n("KButton"),p=n("KEmptyState");return K(),f(p,null,{title:t(()=>[a(r,{class:"mb-3",icon:"dangerCircleOutline",size:"42"}),e(),o("p",null,k(c.productName)+" is running in Standalone mode.",1)]),message:t(()=>[N]),cta:t(()=>[a(i,{to:`https://kuma.io/docs/${s.kumaDocsVersion}/documentation/deployments/`,target:"_blank",appearance:"primary"},{default:t(()=>[e(`
        Learn More
      `)]),_:1},8,["to"])]),_:1})}const I=g(y,[["render",B]]);export{I as M};
