import{d as h,G as d,e as g,o as l,p as u,w as i,c as k,J as b,K as y,F as m,b as w,t as p,l as T,n as B,_ as x}from"./index-CFsM3b-2.js";const C=h({__name:"TagList",props:{tags:{},shouldTruncate:{type:Boolean,default:!1},alignment:{default:"left"}},setup(v){const s=v,c=d(()=>(Array.isArray(s.tags)?s.tags:Object.entries(s.tags??{}).map(([n,a])=>({label:n,value:a}))).map(n=>{const{label:a,value:t}=n,r=f(n),_=a.includes(".kuma.io/")||a.startsWith("kuma.io/");return{label:a,value:t,route:r,isKuma:_}})),o=d(()=>s.shouldTruncate||Object.keys(c.value).length>10);function f(e){if(e.value!=="*")switch(e.label){case"kuma.io/zone":return{name:"data-plane-list-view",query:{s:`zone:${e.value}`}};case"kuma.io/service":return{name:"data-plane-list-view",query:{s:`service:${e.value}`}};case"kuma.io/mesh":return{name:"mesh-detail-view",params:{mesh:e.value}};default:return}}return(e,n)=>{const a=g("XBadge");return l(),u(m(o.value?"KTruncate":"div"),{width:o.value?"auto":void 0,class:B({"tag-list":!o.value,"tag-list--align-right":s.alignment==="right"})},{default:i(()=>[(l(!0),k(b,null,y(c.value,(t,r)=>(l(),u(a,{key:r,class:"tag",appearance:t.isKuma?"info":"neutral"},{default:i(()=>[(l(),u(m(t.route?"XAction":"span"),{to:t.route},{default:i(()=>[w(p(t.label)+":",1),T("b",null,p(t.value),1)]),_:2},1032,["to"]))]),_:2},1032,["appearance"]))),128))]),_:1},8,["width","class"])}}}),L=x(C,[["__scopeId","data-v-df99d86a"]]);export{L as T};
