import{d as k,N as _,a as b,o as s,b as c,w as d,c as m,F as p,B as w,G as f,f as h,t as v,p as B,n as L,_ as T}from"./index-079a3a85.js";function x(o){return Object.entries(o??{}).map(([e,l])=>({label:e,value:l}))}const K=k({__name:"TagList",props:{tags:{},shouldTruncate:{type:Boolean,default:!1},hideLabelKey:{type:Boolean,default:!1},alignment:{default:"left"}},setup(o){const e=o,l=_(()=>(Array.isArray(e.tags)?e.tags:x(e.tags)).map(u=>{const{label:n,value:a}=u,i=y(u),g=n.includes(".kuma.io/")||n.startsWith("kuma.io/");return{label:n,value:a,route:i,isKuma:g}})),r=_(()=>e.shouldTruncate||Object.keys(l.value).length>10);function y(t){if(t.value!=="*")try{switch(t.label){case"kuma.io/zone":return{name:"zone-cp-detail-view",params:{zone:t.value}};case"kuma.io/service":return{name:"service-detail-view",params:{service:t.value}};case"kuma.io/mesh":return{name:"mesh-detail-view",params:{mesh:t.value}};default:return}}catch{return}}return(t,u)=>{const n=b("KBadge");return s(),c(f(r.value?"KTruncate":"div"),{width:r.value?"auto":void 0,class:L({"tag-list":!r.value,"tag-list--align-right":e.alignment==="right"})},{default:d(()=>[(s(!0),m(p,null,w(l.value,(a,i)=>(s(),c(n,{key:i,"max-width":"auto",class:"tag",appearance:a.isKuma?"default":"neutral"},{default:d(()=>[(s(),c(f(a.route?"RouterLink":"span"),{to:a.route},{default:d(()=>[e.hideLabelKey?(s(),m(p,{key:0},[h(v(a.value),1)],64)):(s(),m(p,{key:1},[h(v(a.label)+":",1),B("b",null,v(a.value),1)],64))]),_:2},1032,["to"]))]),_:2},1032,["appearance"]))),128))]),_:1},8,["width","class"])}}});const z=T(K,[["__scopeId","data-v-599273de"]]);export{z as T};
