import{d,u as _,c as n,r as p,o as i,a as f,x as N,v,w as u,h as b,g as h,t as x,b as k,af as y}from"./index-f1b8ae6a.js";import{f as T}from"./RouteView.vue_vue_type_script_setup_true_lang-4a32e1ca.js";const g=d({__name:"NavTabs",props:{tabs:{type:Array,required:!0}},setup(c){const a=c,r=_(),m=n(()=>a.tabs.map(t=>({title:t.title,hash:"#"+t.routeName}))),l=n(()=>{const t=r.matched.map(e=>e.meta.module??"").filter(e=>e!=="");t.reverse();const s=a.tabs.find(e=>!!(e.routeName===r.name||t.includes(e.module)));return"#"+((s==null?void 0:s.routeName)??a.tabs[0].routeName)});return(t,s)=>{const o=p("router-link");return i(),f(k(y),{tabs:m.value,"model-value":l.value,"has-panels":!1,class:"nav-tabs","data-testid":"nav-tabs"},N({_:2},[v(a.tabs,e=>({name:`${e.routeName}-anchor`,fn:u(()=>[b(o,{to:{name:e.routeName}},{default:u(()=>[h(x(e.title),1)]),_:2},1032,["to"])])}))]),1032,["tabs","model-value"])}}});const w=T(g,[["__scopeId","data-v-1c3c46ad"]]);export{w as N};
