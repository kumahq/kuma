import{d as h,M as d,a as g,o,b as c,w as m,c as k,F as b,I as y,D as p,f as w,t as v,p as T,n as x,_ as B}from"./index-8567ed34.js";function L(r){return Object.entries(r??{}).map(([e,n])=>({label:e,value:n}))}const C=h({__name:"TagList",props:{tags:{},shouldTruncate:{type:Boolean,default:!1},alignment:{default:"left"}},setup(r){const e=r,n=d(()=>(Array.isArray(e.tags)?e.tags:L(e.tags)).map(u=>{const{label:s,value:t}=u,i=_(u),f=s.includes(".kuma.io/")||s.startsWith("kuma.io/");return{label:s,value:t,route:i,isKuma:f}})),l=d(()=>e.shouldTruncate||Object.keys(n.value).length>10);function _(a){if(a.value!=="*")try{switch(a.label){case"kuma.io/zone":return{name:"zone-cp-detail-view",params:{zone:a.value}};case"kuma.io/service":return{name:"service-detail-view",params:{service:a.value}};case"kuma.io/mesh":return{name:"mesh-detail-view",params:{mesh:a.value}};default:return}}catch{return}}return(a,u)=>{const s=g("KBadge");return o(),c(p(l.value?"KTruncate":"div"),{width:l.value?"auto":void 0,class:x({"tag-list":!l.value,"tag-list--align-right":e.alignment==="right"})},{default:m(()=>[(o(!0),k(b,null,y(n.value,(t,i)=>(o(),c(s,{key:i,"max-width":"auto",class:"tag",appearance:t.isKuma?"info":"neutral"},{default:m(()=>[(o(),c(p(t.route?"RouterLink":"span"),{to:t.route},{default:m(()=>[w(v(t.label)+":",1),T("b",null,v(t.value),1)]),_:2},1032,["to"]))]),_:2},1032,["appearance"]))),128))]),_:1},8,["width","class"])}}});const z=B(C,[["__scopeId","data-v-625f3123"]]);export{z as T};
