import{d as X,r as m,o as c,q as d,w as e,b as r,m as p,e as a,t as l,T as F,N,c as C,U as g,s as u}from"./index-BkjofXKI.js";import{_ as $}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-DZEpq0tr.js";const A={key:0,class:"stack-with-borders","data-testid":"structured-view"},B={key:1,class:"mt-4"},L=X({__name:"HostnameGeneratorSummaryView",props:{items:{}},setup(w){const k=w;return(M,s)=>{const x=m("RouteTitle"),f=m("XAction"),z=m("XSelect"),h=m("XLayout"),S=m("DataSource"),b=m("AppView"),R=m("DataCollection"),V=m("RouteView");return c(),d(V,{name:"hostname-generator-summary-view",params:{name:"",codeSearch:"",codeFilter:!1,codeRegExp:!1,format:"structured"}},{default:e(({route:o,t:i,can:v})=>[r(R,{items:k.items,predicate:n=>n.id===o.params.name},{item:e(({item:n})=>[r(b,null,{title:e(()=>[p("h2",null,[r(f,{to:{name:"hostname-generator-detail-view",params:{name:o.params.name}}},{default:e(()=>[r(x,{title:i("hostname-generators.routes.item.title",{name:n.name})},null,8,["title"])]),_:2},1032,["to"])])]),default:e(()=>[s[7]||(s[7]=a()),r(h,{type:"stack"},{default:e(()=>[p("header",null,[r(h,{type:"separated",size:"max"},{default:e(()=>[p("h3",null,l(i("hostname-generators.routes.item.config")),1),s[0]||(s[0]=a()),p("div",null,[r(z,{label:i("hostname-generators.routes.item.format"),selected:o.params.format,onChange:t=>{o.update({format:t})}},F({_:2},[N(["structured","yaml"],t=>({name:`${t}-option`,fn:e(()=>[a(l(i(`hostname-generators.routes.item.formats.${t}`)),1)])}))]),1032,["label","selected","onChange"])])]),_:2},1024)]),s[6]||(s[6]=a()),o.params.format==="structured"?(c(),C("div",A,[n.namespace.length>0?(c(),d(g,{key:0,layout:"horizontal"},{title:e(()=>[a(l(i("hostname-generators.common.namespace")),1)]),body:e(()=>[a(l(n.namespace),1)]),_:2},1024)):u("",!0),s[4]||(s[4]=a()),v("use zones")&&n.zone?(c(),d(g,{key:1,layout:"horizontal"},{title:e(()=>[a(l(i("hostname-generators.common.zone")),1)]),body:e(()=>[r(f,{to:{name:"zone-cp-detail-view",params:{zone:n.zone}}},{default:e(()=>[a(l(n.zone),1)]),_:2},1032,["to"])]),_:2},1024)):u("",!0),s[5]||(s[5]=a()),n.spec.template?(c(),d(g,{key:2,layout:"horizontal"},{title:e(()=>[a(l(i("hostname-generators.common.template")),1)]),body:e(()=>[a(l(n.spec.template),1)]),_:2},1024)):u("",!0)])):(c(),C("div",B,[r($,{resource:n.$raw,"is-searchable":"",query:o.params.codeSearch,"is-filter-mode":o.params.codeFilter,"is-reg-exp-mode":o.params.codeRegExp,onQueryChange:t=>o.update({codeSearch:t}),onFilterModeChange:t=>o.update({codeFilter:t}),onRegExpModeChange:t=>o.update({codeRegExp:t})},{default:e(({copy:t,copying:E})=>[E?(c(),d(S,{key:0,src:`/hostname-generators/${o.params.name}/as/kubernetes?no-store`,onChange:_=>{t(y=>y(_))},onError:_=>{t((y,D)=>D(_))}},null,8,["src","onChange","onError"])):u("",!0)]),_:2},1032,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]))]),_:2},1024)]),_:2},1024)]),_:2},1032,["items","predicate"])]),_:1})}}});export{L as default};
