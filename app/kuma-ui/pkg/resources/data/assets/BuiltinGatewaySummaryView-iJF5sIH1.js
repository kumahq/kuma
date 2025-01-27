import{d as B,r as s,o as l,q as p,w as e,b as n,ao as F,m as c,t as d,e as i,c as _,M as D,N as h,T as M,U as $,s as C,_ as A}from"./index-oTPgN0we.js";import{_ as G}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-CawNJ-kc.js";const T={key:0,class:"stack-with-borders","data-testid":"structured-view"},q={key:1,class:"mt-4"},L=B({__name:"BuiltinGatewaySummaryView",props:{items:{},routeName:{}},setup(x){const g=x;return(z,o)=>{const S=s("XEmptyState"),k=s("RouteTitle"),b=s("XAction"),E=s("XSelect"),f=s("XLayout"),V=s("DataSource"),R=s("AppView"),X=s("RouteView");return l(),p(X,{name:g.routeName,params:{mesh:"",gateway:"",codeSearch:"",codeFilter:!1,codeRegExp:!1,format:"structured"}},{default:e(({route:a,t:r})=>[n(F,{items:g.items,predicate:u=>u.id===a.params.gateway,find:!0},{empty:e(()=>[n(S,null,{title:e(()=>[c("h2",null,d(r("common.collection.summary.empty_title",{type:"Gateway"})),1)]),default:e(()=>[o[0]||(o[0]=i()),c("p",null,d(r("common.collection.summary.empty_message",{type:"Gateway"})),1)]),_:2},1024)]),default:e(({items:u})=>[(l(!0),_(D,null,h([u[0]],m=>(l(),p(R,{key:m.id},{title:e(()=>[c("h2",null,[n(b,{to:{name:"builtin-gateway-detail-view",params:{mesh:m.mesh,gateway:m.id}}},{default:e(()=>[n(k,{title:r("builtin-gateways.routes.item.title",{name:m.name})},null,8,["title"])]),_:2},1032,["to"])])]),default:e(()=>[o[4]||(o[4]=i()),n(f,{type:"stack"},{default:e(()=>[c("header",null,[n(f,{type:"separated",size:"max"},{default:e(()=>[c("h3",null,d(r("gateways.routes.item.config")),1),o[1]||(o[1]=i()),c("div",null,[n(E,{label:r("gateways.routes.item.format"),selected:a.params.format,onChange:t=>{a.update({format:t})}},M({_:2},[h(["structured","yaml"],t=>({name:`${t}-option`,fn:e(()=>[i(d(r(`gateways.routes.item.formats.${t}`)),1)])}))]),1032,["label","selected","onChange"])])]),_:2},1024)]),o[3]||(o[3]=i()),a.params.format==="structured"?(l(),_("div",T,[m.namespace.length>0?(l(),p($,{key:0,layout:"horizontal"},{title:e(()=>[i(d(r("gateways.routes.item.namespace")),1)]),body:e(()=>[i(d(m.namespace),1)]),_:2},1024)):C("",!0)])):(l(),_("div",q,[n(G,{resource:m.config,"is-searchable":"",query:a.params.codeSearch,"is-filter-mode":a.params.codeFilter,"is-reg-exp-mode":a.params.codeRegExp,onQueryChange:t=>a.update({codeSearch:t}),onFilterModeChange:t=>a.update({codeFilter:t}),onRegExpModeChange:t=>a.update({codeRegExp:t})},{default:e(({copy:t,copying:v})=>[v?(l(),p(V,{key:0,src:`/meshes/${a.params.mesh}/mesh-gateways/${a.params.gateway}/as/kubernetes?no-store`,onChange:y=>{t(w=>w(y))},onError:y=>{t((w,N)=>N(y))}},null,8,["src","onChange","onError"])):C("",!0)]),_:2},1032,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]))]),_:2},1024)]),_:2},1024))),128))]),_:2},1032,["items","predicate"])]),_:1},8,["name"])}}}),U=A(L,[["__scopeId","data-v-6be11ac1"]]);export{U as default};
