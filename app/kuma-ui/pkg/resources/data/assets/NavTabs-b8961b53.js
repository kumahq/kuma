import{d as m,e as d,h as n,r as i,o as p,i as v,a8 as N,I as f,w as u,j as b,n as h,H as x,k as y,al as k,t as T}from"./index-71437355.js";const L=m({__name:"NavTabs",props:{tabs:{type:Array,required:!0}},setup(c){const a=c,o=d(),_=n(()=>a.tabs.map(t=>({title:t.title,hash:"#"+t.routeName}))),l=n(()=>{const t=o.matched.map(e=>e.meta.module??"").filter(e=>e!=="");t.reverse();const s=a.tabs.find(e=>!!(e.routeName===o.name||t.includes(e.module)));return"#"+((s==null?void 0:s.routeName)??a.tabs[0].routeName)});return(t,s)=>{const r=i("RouterLink");return p(),v(y(k),{tabs:_.value,"model-value":l.value,"has-panels":!1,class:"nav-tabs","data-testid":"nav-tabs"},N({_:2},[f(a.tabs,e=>({name:`${e.routeName}-anchor`,fn:u(()=>[b(r,{"data-testid":`${e.routeName}-tab`,to:{name:e.routeName}},{default:u(()=>[h(x(e.title),1)]),_:2},1032,["data-testid","to"])])}))]),1032,["tabs","model-value"])}}});const C=T(L,[["__scopeId","data-v-efa5cb58"]]);export{C as N};
