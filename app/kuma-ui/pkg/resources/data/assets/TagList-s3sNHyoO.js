import{d as b,k as g,I as p,a as k,o as n,b as c,w as m,F as w,C as v,t as f,f as y,m as T,E as B,c as C,n as x,_ as L}from"./index-ENpHoh_M.js";function z(r){return Object.entries(r??{}).map(([e,o])=>({label:e,value:o}))}const K=b({__name:"TagList",props:{tags:{},shouldTruncate:{type:Boolean,default:!1},alignment:{default:"left"}},setup(r){const e=r,o=g(),d=p(()=>(Array.isArray(e.tags)?e.tags:z(e.tags)).map(l=>{const{label:s,value:t}=l,i=h(l),_=s.includes(".kuma.io/")||s.startsWith("kuma.io/");return{label:s,value:t,route:i,isKuma:_}})),u=p(()=>e.shouldTruncate||Object.keys(d.value).length>10);function h(a){if(a.value!=="*")try{switch(a.label){case"kuma.io/zone":return o("use zones")?{name:"zone-cp-detail-view",params:{zone:a.value}}:void 0;case"kuma.io/service":return{name:"service-detail-view",params:{service:a.value}};case"kuma.io/mesh":return{name:"mesh-detail-view",params:{mesh:a.value}};default:return}}catch{return}}return(a,l)=>{const s=k("KBadge");return n(),c(v(u.value?"KTruncate":"div"),{width:u.value?"auto":void 0,class:x({"tag-list":!u.value,"tag-list--align-right":e.alignment==="right"})},{default:m(()=>[(n(!0),C(w,null,B(d.value,(t,i)=>(n(),c(s,{key:i,"max-width":"auto",class:"tag",appearance:t.isKuma?"info":"neutral"},{default:m(()=>[(n(),c(v(t.route?"RouterLink":"span"),{to:t.route},{default:m(()=>[y(f(t.label)+":",1),T("b",null,f(t.value),1)]),_:2},1032,["to"]))]),_:2},1032,["appearance"]))),128))]),_:1},8,["width","class"])}}}),A=L(K,[["__scopeId","data-v-5cb3bb02"]]);export{A as T};
