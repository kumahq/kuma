import{A as m,Q as u,g as l,cl as _,co as d,H as g,i as o,o as K,c as f,w as t,a,b as e,l as n,t as y}from"./index.e014f0d3.js";const k={name:"MultizoneInfo",components:{KButton:m,KEmptyState:u,KIcon:l},data(){return{productName:_}},computed:{...d({kumaDocsVersion:"config/getKumaDocsVersion"})}},B=n("p",null,[e(`
        To access this page, you must be running in `),n("strong",null,"Multi-Zone"),e(` mode.
      `)],-1);function M(s,N,V,h,c,D){const r=o("KIcon"),i=o("KButton"),p=o("KEmptyState");return K(),f(p,null,{title:t(()=>[a(r,{class:"mb-3",icon:"dangerCircleOutline",size:"42"}),e(),n("p",null,y(c.productName)+" is running in Standalone mode.",1)]),message:t(()=>[B]),cta:t(()=>[a(i,{to:`https://kuma.io/docs/${s.kumaDocsVersion}/documentation/deployments/`,target:"_blank",appearance:"primary"},{default:t(()=>[e(`
        Learn More
      `)]),_:1},8,["to"])]),_:1})}const S=g(k,[["render",M]]);export{S as M};
