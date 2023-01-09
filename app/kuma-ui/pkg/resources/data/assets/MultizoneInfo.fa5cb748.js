import{O as i,Q as u,m as l,cl as _,co as d,G as g,h as o,o as K,c as f,w as t,a,b as e,j as n,t as y}from"./index.7a1b9f3b.js";const h={name:"MultizoneInfo",components:{KButton:i,KEmptyState:u,KIcon:l},data(){return{productName:_}},computed:{...d({kumaDocsVersion:"config/getKumaDocsVersion"})}},k=n("p",null,[e(`
        To access this page, you must be running in `),n("strong",null,"Multi-Zone"),e(` mode.
      `)],-1);function B(s,M,N,V,c,D){const r=o("KIcon"),m=o("KButton"),p=o("KEmptyState");return K(),f(p,null,{title:t(()=>[a(r,{class:"mb-3",icon:"dangerCircleOutline",size:"42"}),e(),n("p",null,y(c.productName)+" is running in Standalone mode.",1)]),message:t(()=>[k]),cta:t(()=>[a(m,{to:`https://kuma.io/docs/${s.kumaDocsVersion}/documentation/deployments/`,target:"_blank",appearance:"primary"},{default:t(()=>[e(`
        Learn More
      `)]),_:1},8,["to"])]),_:1})}const S=g(h,[["render",B]]);export{S as M};
