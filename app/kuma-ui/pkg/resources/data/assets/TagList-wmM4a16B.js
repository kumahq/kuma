import{d as _,L as m,r as g,o,m as u,w as i,c as k,F as w,s as y,E as d,e as b,t as p,k as T,n as B,q as x}from"./index--1DEc0sn.js";const L=_({__name:"TagList",props:{tags:{},shouldTruncate:{type:Boolean,default:!1},alignment:{default:"left"}},setup(v){const s=v,c=m(()=>(Array.isArray(s.tags)?s.tags:Object.entries(s.tags??{}).map(([n,a])=>({label:n,value:a}))).map(n=>{const{label:a,value:t}=n,l=h(n),f=a.includes(".kuma.io/")||a.startsWith("kuma.io/");return{label:a,value:t,route:l,isKuma:f}})),r=m(()=>s.shouldTruncate||Object.keys(c.value).length>10);function h(e){if(e.value!=="*")switch(e.label){case"kuma.io/zone":return{name:"data-plane-list-view",query:{s:`zone:${e.value}`}};case"kuma.io/service":return{name:"data-plane-list-view",query:{s:`service:${e.value}`}};case"kuma.io/mesh":return{name:"mesh-detail-view",params:{mesh:e.value}};default:return}}return(e,n)=>{const a=g("KBadge");return o(),u(d(r.value?"KTruncate":"div"),{width:r.value?"auto":void 0,class:B({"tag-list":!r.value,"tag-list--align-right":s.alignment==="right"})},{default:i(()=>[(o(!0),k(w,null,y(c.value,(t,l)=>(o(),u(a,{key:l,"max-width":"auto",class:"tag",appearance:t.isKuma?"info":"neutral"},{default:i(()=>[(o(),u(d(t.route?"RouterLink":"span"),{to:t.route},{default:i(()=>[b(p(t.label)+":",1),T("b",null,p(t.value),1)]),_:2},1032,["to"]))]),_:2},1032,["appearance"]))),128))]),_:1},8,["width","class"])}}}),K=x(L,[["__scopeId","data-v-283453ac"]]);export{K as T};
