import{d as F,r as l,o as i,q as m,w as t,b as r,m as c,e as s,t as d,Q as L,L as M,c as h,R as u,s as f}from"./index-CKQWVGYP.js";import{_ as N}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-BDFzF6sj.js";const $={key:0,class:"stack-with-borders","data-testid":"structured-view"},A={key:1,class:"mt-4"},Q=F({__name:"MeshExternalServiceSummaryView",props:{items:{}},setup(C){const x=C;return(T,e)=>{const b=l("RouteTitle"),y=l("XAction"),k=l("XSelect"),g=l("XLayout"),w=l("KumaPort"),S=l("XBadge"),z=l("DataSource"),E=l("AppView"),R=l("DataCollection"),V=l("RouteView");return i(),m(V,{name:"mesh-external-service-summary-view",params:{mesh:"",service:"",codeSearch:"",codeFilter:!1,codeRegExp:!1,format:"structured"}},{default:t(({route:a,t:p,can:X})=>[r(R,{items:x.items,predicate:n=>n.id===a.params.service},{item:t(({item:n})=>[r(E,null,{title:t(()=>[c("h2",null,[r(y,{to:{name:"mesh-external-service-detail-view",params:{mesh:a.params.mesh,service:a.params.service}}},{default:t(()=>[r(b,{title:p("services.routes.item.title",{name:n.name})},null,8,["title"])]),_:2},1032,["to"])])]),default:t(()=>[e[13]||(e[13]=s()),r(g,{type:"stack"},{default:t(()=>[c("header",null,[r(g,{type:"separated",size:"max"},{default:t(()=>[c("h3",null,d(p("services.routes.item.config")),1),e[0]||(e[0]=s()),c("div",null,[r(k,{label:p("services.routes.item.format"),selected:a.params.format,onChange:o=>{a.update({format:o})}},L({_:2},[M(["structured","yaml"],o=>({name:`${o}-option`,fn:t(()=>[s(d(p(`services.routes.item.formats.${o}`)),1)])}))]),1032,["label","selected","onChange"])])]),_:2},1024)]),e[12]||(e[12]=s()),a.params.format==="structured"?(i(),h("div",$,[n.namespace.length>0?(i(),m(u,{key:0,layout:"horizontal"},{title:t(()=>e[1]||(e[1]=[s(`
                    Namespace
                  `)])),body:t(()=>[s(d(n.namespace),1)]),_:2},1024)):f("",!0),e[9]||(e[9]=s()),X("use zones")&&n.zone?(i(),m(u,{key:1,layout:"horizontal"},{title:t(()=>e[3]||(e[3]=[s(`
                    Zone
                  `)])),body:t(()=>[r(y,{to:{name:"zone-cp-detail-view",params:{zone:n.zone}}},{default:t(()=>[s(d(n.zone),1)]),_:2},1032,["to"])]),_:2},1024)):f("",!0),e[10]||(e[10]=s()),n.spec.match?(i(),m(u,{key:2,layout:"horizontal"},{title:t(()=>e[5]||(e[5]=[s(`
                    Port
                  `)])),body:t(()=>[r(w,{port:n.spec.match},null,8,["port"])]),_:2},1024)):f("",!0),e[11]||(e[11]=s()),r(u,{layout:"horizontal"},{title:t(()=>e[7]||(e[7]=[s(`
                    TLS
                  `)])),body:t(()=>[r(S,{appearance:"neutral"},{default:t(()=>{var o;return[s(d((o=n.spec.tls)!=null&&o.enabled?"Enabled":"Disabled"),1)]}),_:2},1024)]),_:2},1024)])):(i(),h("div",A,[r(N,{resource:n.config,"is-searchable":"",query:a.params.codeSearch,"is-filter-mode":a.params.codeFilter,"is-reg-exp-mode":a.params.codeRegExp,onQueryChange:o=>a.update({codeSearch:o}),onFilterModeChange:o=>a.update({codeFilter:o}),onRegExpModeChange:o=>a.update({codeRegExp:o})},{default:t(({copy:o,copying:D})=>[D?(i(),m(z,{key:0,src:`/meshes/${a.params.mesh}/mesh-service/${a.params.service}/as/kubernetes?no-store`,onChange:_=>{o(v=>v(_))},onError:_=>{o((v,B)=>B(_))}},null,8,["src","onChange","onError"])):f("",!0)]),_:2},1032,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]))]),_:2},1024)]),_:2},1024)]),_:2},1032,["items","predicate"])]),_:1})}}});export{Q as default};
