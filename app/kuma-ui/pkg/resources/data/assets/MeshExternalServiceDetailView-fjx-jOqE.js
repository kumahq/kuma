import{d as D,e as s,o as d,m as l,w as a,a as p,k as g,P as c,b as o,t as _,p as i,q as R}from"./index-C_eW3RRu.js";import{_ as z}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-BajQNyhd.js";const M={class:"stack"},B={class:"columns"},F=D({__name:"MeshExternalServiceDetailView",props:{data:{}},setup(C){const n=C;return(m,e)=>{const y=s("XAction"),x=s("KumaPort"),v=s("XBadge"),b=s("KCard"),h=s("DataSource"),k=s("AppView"),w=s("RouteView");return d(),l(w,{name:"mesh-external-service-detail-view",params:{codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:a(({route:r,can:E})=>[p(k,null,{default:a(()=>[g("div",M,[p(b,null,{default:a(()=>[g("div",B,[n.data.namespace.length>0?(d(),l(c,{key:0},{title:a(()=>e[0]||(e[0]=[o(`
                Namespace
              `)])),body:a(()=>[o(_(n.data.namespace),1)]),_:1})):i("",!0),e[8]||(e[8]=o()),E("use zones")&&n.data.zone?(d(),l(c,{key:1},{title:a(()=>e[2]||(e[2]=[o(`
                Zone
              `)])),body:a(()=>[p(y,{to:{name:"zone-cp-detail-view",params:{zone:n.data.zone}}},{default:a(()=>[o(_(n.data.zone),1)]),_:1},8,["to"])]),_:1})):i("",!0),e[9]||(e[9]=o()),m.data.spec.match?(d(),l(c,{key:2,class:"port"},{title:a(()=>e[4]||(e[4]=[o(`
                Port
              `)])),body:a(()=>[p(x,{port:m.data.spec.match},null,8,["port"])]),_:1})):i("",!0),e[10]||(e[10]=o()),m.data.spec.match?(d(),l(c,{key:3,class:"tls"},{title:a(()=>e[6]||(e[6]=[o(`
                TLS
              `)])),body:a(()=>[p(v,{appearance:"neutral"},{default:a(()=>{var t;return[o(_((t=m.data.spec.tls)!=null&&t.enabled?"Enabled":"Disabled"),1)]}),_:1})]),_:1})):i("",!0)])]),_:2},1024),e[11]||(e[11]=o()),p(z,{resource:n.data.config,"is-searchable":"",query:r.params.codeSearch,"is-filter-mode":r.params.codeFilter,"is-reg-exp-mode":r.params.codeRegExp,onQueryChange:t=>r.update({codeSearch:t}),onFilterModeChange:t=>r.update({codeFilter:t}),onRegExpModeChange:t=>r.update({codeRegExp:t})},{default:a(({copy:t,copying:V})=>[V?(d(),l(h,{key:0,src:`/meshes/${n.data.mesh}/mesh-external-service/${n.data.id}/as/kubernetes?no-store`,onChange:u=>{t(f=>f(u))},onError:u=>{t((f,S)=>S(u))}},null,8,["src","onChange","onError"])):i("",!0)]),_:2},1032,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])])]),_:2},1024)]),_:1})}}}),K=R(F,[["__scopeId","data-v-3272a48f"]]);export{K as default};
