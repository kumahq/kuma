import{d as _,L as d,r as g,o as r,m as c,w as m,c as k,F as b,s as w,E as p,e as y,t as v,k as T,n as B,q as L}from"./index-Be4yFAuI.js";function x(o){return Object.entries(o??{}).map(([e,n])=>({label:e,value:n}))}const C=_({__name:"TagList",props:{tags:{},shouldTruncate:{type:Boolean,default:!1},alignment:{default:"left"}},setup(o){const e=o,n=d(()=>(Array.isArray(e.tags)?e.tags:x(e.tags)).map(u=>{const{label:s,value:t}=u,i=f(u),h=s.includes(".kuma.io/")||s.startsWith("kuma.io/");return{label:s,value:t,route:i,isKuma:h}})),l=d(()=>e.shouldTruncate||Object.keys(n.value).length>10);function f(a){if(a.value!=="*")switch(a.label){case"kuma.io/zone":return{name:"data-plane-list-view",query:{s:`zone:${a.value}`}};case"kuma.io/service":return{name:"data-plane-list-view",query:{s:`service:${a.value}`}};case"kuma.io/mesh":return{name:"mesh-detail-view",params:{mesh:a.value}};default:return}}return(a,u)=>{const s=g("KBadge");return r(),c(p(l.value?"KTruncate":"div"),{width:l.value?"auto":void 0,class:B({"tag-list":!l.value,"tag-list--align-right":e.alignment==="right"})},{default:m(()=>[(r(!0),k(b,null,w(n.value,(t,i)=>(r(),c(s,{key:i,"max-width":"auto",class:"tag",appearance:t.isKuma?"info":"neutral"},{default:m(()=>[(r(),c(p(t.route?"RouterLink":"span"),{to:t.route},{default:m(()=>[y(v(t.label)+":",1),T("b",null,v(t.value),1)]),_:2},1032,["to"]))]),_:2},1032,["appearance"]))),128))]),_:1},8,["width","class"])}}}),q=L(C,[["__scopeId","data-v-3d970d4e"]]);export{q as T};