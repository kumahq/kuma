import{d as _,c as m,h as g,o as r,a as c,w as d,b as k,F as b,B as w,r as p,k as y,t as v,g as B,l as T,A as x}from"./index-ChMk9xbI.js";function L(l){return Object.entries(l??{}).map(([e,n])=>({label:e,value:n}))}const C=_({__name:"TagList",props:{tags:{},shouldTruncate:{type:Boolean,default:!1},alignment:{default:"left"}},setup(l){const e=l,n=m(()=>(Array.isArray(e.tags)?e.tags:L(e.tags)).map(u=>{const{label:s,value:t}=u,i=h(u),f=s.includes(".kuma.io/")||s.startsWith("kuma.io/");return{label:s,value:t,route:i,isKuma:f}})),o=m(()=>e.shouldTruncate||Object.keys(n.value).length>10);function h(a){if(a.value!=="*")switch(a.label){case"kuma.io/zone":return{name:"data-plane-list-view",query:{s:`zone:${a.value}`}};case"kuma.io/service":return{name:"data-plane-list-view",query:{s:`service:${a.value}`}};case"kuma.io/mesh":return{name:"mesh-detail-view",params:{mesh:a.value}};default:return}}return(a,u)=>{const s=g("KBadge");return r(),c(p(o.value?"KTruncate":"div"),{width:o.value?"auto":void 0,class:T({"tag-list":!o.value,"tag-list--align-right":e.alignment==="right"})},{default:d(()=>[(r(!0),k(b,null,w(n.value,(t,i)=>(r(),c(s,{key:i,"max-width":"auto",class:"tag",appearance:t.isKuma?"info":"neutral"},{default:d(()=>[(r(),c(p(t.route?"RouterLink":"span"),{to:t.route},{default:d(()=>[y(v(t.label)+":",1),B("b",null,v(t.value),1)]),_:2},1032,["to"]))]),_:2},1032,["appearance"]))),128))]),_:1},8,["width","class"])}}}),z=x(C,[["__scopeId","data-v-3d970d4e"]]);export{z as T};
