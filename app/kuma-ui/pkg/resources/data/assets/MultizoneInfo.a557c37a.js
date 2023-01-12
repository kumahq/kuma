import{O as m,p as u,m as l,bW as d,c6 as _,k as g,cd as n,o as K,i as f,w as e,a,b as t,e as o,bV as k}from"./index.2d238a59.js";const y={name:"MultizoneInfo",components:{KButton:m,KEmptyState:u,KIcon:l},data(){return{productName:d}},computed:{..._({kumaDocsVersion:"config/getKumaDocsVersion"})}},V=o("p",null,[t(`
        To access this page, you must be running in `),o("strong",null,"Multi-Zone"),t(` mode.
      `)],-1);function b(s,B,M,N,c,h){const r=n("KIcon"),p=n("KButton"),i=n("KEmptyState");return K(),f(i,null,{title:e(()=>[a(r,{class:"mb-3",icon:"dangerCircleOutline",size:"42"}),t(),o("p",null,k(c.productName)+" is running in Standalone mode.",1)]),message:e(()=>[V]),cta:e(()=>[a(p,{to:`https://kuma.io/docs/${s.kumaDocsVersion}/documentation/deployments/`,target:"_blank",appearance:"primary"},{default:e(()=>[t(`
        Learn More
      `)]),_:1},8,["to"])]),_:1})}const I=g(y,[["render",b]]);export{I as M};
