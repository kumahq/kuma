import{d as q,g as x,e as H,h as E,o as a,l as i,p as t,n,H as s,k as e,j as r,Q as de,y as k,az as A,r as m,i as C,w as l,F as h,m as b,I as ue,X as me,aA as M,aB as X,aC as j,K as G,$ as _e,aD as pe,aE as fe,aF as ve,t as be}from"./index-bc0f9b6f.js";import{_ as g}from"./CodeBlock.vue_vue_type_style_index_0_lang-dcf424de.js";const ze={class:"form-step-title"},ke=t("span",{class:"form-step-number"},"1",-1),he={class:"instruction-list"},ge={href:"https://helm.sh/docs/intro/install/"},ye={class:"form-step-title"},Ce=t("span",{class:"form-step-number"},"2",-1),Ee=t("p",null,"On your local machine, create a namespace in your Kubernetes cluster and pull down the kong Helm repo.",-1),Ke={class:"instruction-list"},$e={class:"form-step-title"},Ve=t("span",{class:"form-step-number"},"3",-1),Ie={class:"form-step-title"},we=t("span",{class:"form-step-number"},"4",-1),Te={class:"field-group-label mt-4"},Ne={class:"mt-4"},Se=q({__name:"ZoneCreateKubernetesInstructions",props:{zoneName:{type:String,required:!0},globalKdsAddress:{type:String,required:!0},zoneIngressEnabled:{type:Boolean,required:!0},zoneEgressEnabled:{type:Boolean,required:!0},token:{type:String,required:!0},base64EncodedToken:{type:String,required:!0}},setup(T){const o=x(),y=H(),_=T,$=E(()=>o.t("zones.form.kubernetes.secret.createSecretCommand",{token:_.base64EncodedToken}).trim()),v=E(()=>{const p={zoneName:_.zoneName,globalKdsAddress:_.globalKdsAddress,zoneIngressEnabled:String(_.zoneIngressEnabled),zoneEgressEnabled:String(_.zoneEgressEnabled)};return typeof y.params.virtualControlPlaneId=="string"&&(p.controlPlaneId=y.params.virtualControlPlaneId),o.t("zones.form.kubernetes.connectZone.config",p).trim()});return(p,V)=>(a(),i("div",null,[t("h3",ze,[ke,n(" "+s(e(o).t("zones.form.kubernetes.prerequisites.title")),1)]),n(),t("ul",he,[t("li",null,[t("b",null,s(e(o).t("zones.form.kubernetes.prerequisites.step1Label"))+s(_.zoneIngressEnabled?" "+e(o).t("zones.form.kubernetes.prerequisites.step1LabelAddendum"):""),1),n(`:
        `+s(e(o).t("zones.form.kubernetes.prerequisites.step1Description",{productName:e(o).t("common.product.name")})),1)]),n(),t("li",null,[t("b",null,s(e(o).t("zones.form.kubernetes.prerequisites.step2Label")),1),n(`:
        `+s(e(o).t("zones.form.kubernetes.prerequisites.step2Description")),1)]),n(),t("li",null,[t("a",ge,s(e(o).t("zones.form.kubernetes.prerequisites.step3LinkTitle")),1),n(" "+s(e(o).t("zones.form.kubernetes.prerequisites.step3Tail")),1)])]),n(),t("h3",ye,[Ce,n(" "+s(e(o).t("zones.form.kubernetes.helm.title")),1)]),n(),Ee,n(),t("ol",Ke,[t("li",null,[t("b",null,s(e(o).t("zones.form.kubernetes.helm.step1Description")),1),n(),r(g,{id:"zone-kubernetes-create-namespace",class:"mt-2",code:e(o).t("zones.form.kubernetes.helm.step1Command"),language:"bash"},null,8,["code"])]),n(),t("li",null,[t("b",null,s(e(o).t("zones.form.kubernetes.helm.step2Description")),1),n(),r(g,{id:"zone-kubernetes-add-charts-repo",class:"mt-2",code:e(o).t("zones.form.kubernetes.helm.step2Command"),language:"bash"},null,8,["code"])]),n(),t("li",null,[t("b",null,s(e(o).t("zones.form.kubernetes.helm.step3Description")),1),n(),r(g,{id:"zone-kubernetes-repo-update",class:"mt-2",code:e(o).t("zones.form.kubernetes.helm.step3Command"),language:"bash"},null,8,["code"])])]),n(),t("h3",$e,[Ve,n(" "+s(e(o).t("zones.form.kubernetes.secret.title")),1)]),n(),t("p",null,s(e(o).t("zones.form.kubernetes.secret.createSecretDescription")),1),n(),r(g,{id:"zone-kubernetes-create-secret",class:"mt-4",code:$.value,language:"bash"},null,8,["code"]),n(),t("h3",Ie,[we,n(" "+s(e(o).t("zones.form.kubernetes.connectZone.title")),1)]),n(),t("p",null,s(e(o).t("zones.form.kubernetes.connectZone.configDescription")),1),n(),t("span",Te,s(e(o).t("zones.form.kubernetes.connectZone.configFileName")),1),n(),r(g,{id:"zone-kubernetes-config-code-block","data-testid":"zone-kubernetes-config",code:v.value,language:"yaml"},null,8,["code"]),n(),t("p",Ne,s(e(o).t("zones.form.kubernetes.connectZone.connectDescription")),1),n(),r(g,{id:"zone-kubernetes-command-code-block",class:"mt-4",code:e(o).t("zones.form.kubernetes.connectZone.connectCommand").trim(),language:"bash"},null,8,["code"])]))}}),Le={class:"form-step-title"},De=t("span",{class:"form-step-number"},"1",-1),Ze={class:"form-step-title"},Ae=t("span",{class:"form-step-number"},"2",-1),qe={class:"field-group-label mt-4"},xe={class:"mt-4"},Be=q({__name:"ZoneCreateUniversalInstructions",props:{zoneName:{type:String,required:!0},globalKdsAddress:{type:String,required:!0},token:{type:String,required:!0}},setup(T){const o=x(),y=H(),_=T,$=E(()=>o.t("zones.form.universal.saveToken.saveTokenCommand",{token:_.token}).trim()),v=E(()=>{const p={zoneName:_.zoneName,globalKdsAddress:_.globalKdsAddress};return typeof y.params.virtualControlPlaneId=="string"&&(p.controlPlaneId=y.params.virtualControlPlaneId),o.t("zones.form.universal.connectZone.config",p).trim()});return(p,V)=>(a(),i("div",null,[t("h3",Le,[De,n(" "+s(e(o).t("zones.form.universal.saveToken.title")),1)]),n(),t("p",null,s(e(o).t("zones.form.universal.saveToken.saveTokenDescription")),1),n(),r(g,{id:"zone-kubernetes-token",class:"mt-4",code:$.value,language:"bash"},null,8,["code"]),n(),t("h3",Ze,[Ae,n(" "+s(e(o).t("zones.form.universal.connectZone.title")),1)]),n(),t("p",null,s(e(o).t("zones.form.universal.connectZone.configDescription")),1),n(),t("span",qe,s(e(o).t("zones.form.universal.connectZone.configFileName")),1),n(),r(g,{id:"zone-universal-config-code-block","data-testid":"zone-universal-config",class:"mt-4",code:v.value,language:"yaml"},null,8,["code"]),n(),t("p",xe,s(e(o).t("zones.form.universal.connectZone.connectDescription")),1),n(),r(g,{id:"zone-universal-connect-command-code-block",class:"mt-4",code:e(o).t("zones.form.universal.connectZone.connectCommand").trim(),language:"bash"},null,8,["code"])]))}}),Re={class:"form-wrapper"},Ue={key:1},Pe={key:2},Oe={class:"form"},Fe={class:"form-header"},Me={class:"form-title"},Xe={class:"text-gradient"},je={key:0},Ge={key:0},He={class:"fact-list"},Qe={class:"form-section"},We={class:"form-section__header"},Je={class:"form-section-title"},Ye={key:0},en={class:"form-section__content"},nn={class:"form-section","data-testid":"connect-zone-instructions"},on={class:"form-section__header"},tn={class:"form-section-title"},sn={key:0},an={class:"form-section__content"},rn={class:"field-group-list"},ln={class:"field-group"},cn={class:"field-group-label"},dn={class:"radio-button-group"},un={class:"field-group"},mn={class:"field-group-label"},_n={class:"radio-button-group"},pn={class:"field-group"},fn={class:"field-group-label"},vn={class:"radio-button-group"},bn={class:"form-section"},zn={class:"form-section__header"},kn={class:"form-section-title"},hn={key:0},gn={class:"form-section__content"},yn={class:"form-section"},Cn={class:"form-section__header"},En={class:"form-section-title"},Kn={key:0},$n={class:"form-section__content"},Vn={key:0},In={class:"mt-2"},wn=q({__name:"CreateView",setup(T){const{t:o,tm:y}=x(),_=de(),$=/^(?![-0-9])[a-z0-9-]{1,63}$/,v=k(null),p=k(!1),V=k(!1),f=k(null),N=k(null),B=k(!1),u=k(""),K=k("kubernetes"),L=k(!0),D=k(!0),I=E(()=>v.value!==null&&v.value.token?v.value.token:""),Q=E(()=>I.value!==""?window.btoa(I.value):""),W=E(()=>u.value===""||p.value||v.value!==null),Z=E(()=>{if(N.value!==null)return N.value;if(f.value instanceof A){const z=f.value.invalidParameters.find(d=>d.field==="name");if(z!==void 0)return z.reason}return null});async function J(){p.value=!0,f.value=null;try{if(!R(u.value))return;v.value=await _.createZone({name:u.value})}catch(z){z instanceof Error?f.value=z:console.error(z)}finally{p.value=!1}}function R(z){const d=$.test(z);return d?N.value=null:N.value=o("zones.create.invalidNameError"),d}function U(){V.value=!V.value}function Y(){B.value=!0}return(z,d)=>{const ee=m("RouteTitle"),w=m("KButton"),ne=m("KAlert"),oe=m("KLabel"),te=m("KInput"),P=m("KRadio"),O=m("KInputSwitch"),F=m("DataSource"),se=m("KEmptyState"),ae=m("KCard"),re=m("KModal"),le=m("AppView"),ie=m("RouteView");return a(),C(ie,{name:"zone-create-view",attrs:{class:"is-fullscreen"}},{default:l(({route:ce})=>[r(le,{fullscreen:!0,breadcrumbs:[]},{title:l(()=>[t("h1",null,[r(ee,{title:e(o)("zones.routes.create.title"),render:!0},null,8,["title"])])]),actions:l(()=>[I.value===""||B.value?(a(),C(w,{key:0,appearance:"outline","data-testid":"exit-button",to:{name:"zone-cp-list-view"}},{default:l(()=>[n(s(e(o)("zones.form.exit")),1)]),_:1})):(a(),C(w,{key:1,appearance:"outline","data-testid":"exit-button",onClick:U},{default:l(()=>[n(s(e(o)("zones.form.exit")),1)]),_:1}))]),default:l(()=>[n(),n(),t("div",Re,[f.value!==null?(a(),C(ne,{key:0,appearance:"danger",class:"mb-4","dismiss-type":"icon","data-testid":"create-zone-error"},{alertMessage:l(()=>[f.value instanceof e(A)&&[409,500].includes(f.value.status)?(a(),i(h,{key:0},[t("p",null,s(e(o)(`zones.create.status_error.${f.value.status}.title`,{name:u.value})),1),n(),t("p",null,s(e(o)(`zones.create.status_error.${f.value.status}.description`)),1)],64)):f.value instanceof e(A)?(a(),i("p",Ue,s(e(o)("common.error_state.api_error",{status:f.value.status,title:f.value.detail})),1)):(a(),i("p",Pe,s(e(o)("common.error_state.default_error")),1))]),_:1})):b("",!0),n(),r(ae,{class:"form-card"},{body:l(()=>[t("div",Oe,[t("div",Fe,[t("div",null,[t("h1",Me,[t("span",Xe,s(e(o)("zones.form.title")),1)]),n(),e(o)("zones.form.description")!==" "?(a(),i("p",je,s(e(o)("zones.form.description")),1)):b("",!0)]),n(),e(y)("zones.form.facts").length>0?(a(),i("div",Ge,[t("ul",He,[(a(!0),i(h,null,ue(e(y)("zones.form.facts"),(c,S)=>(a(),i("li",{key:S,class:"fact-list__item"},[r(e(me),{color:e(M)},null,8,["color"]),n(" "+s(c),1)]))),128))])])):b("",!0)]),n(),t("div",Qe,[t("div",We,[t("h2",Je,s(e(o)("zones.form.section.name.title")),1),n(),e(o)("zones.form.section.name.description")!==" "?(a(),i("p",Ye,s(e(o)("zones.form.section.name.description")),1)):b("",!0)]),n(),t("div",en,[t("div",null,[r(oe,{for:"zone-name",required:"","tooltip-attributes":{placement:"right"}},{tooltip:l(()=>[n(s(e(o)("zones.form.name_tooltip")),1)]),default:l(()=>[n(s(e(o)("zones.form.nameLabel"))+" ",1)]),_:1}),n(),r(te,{id:"zone-name",modelValue:u.value,"onUpdate:modelValue":d[0]||(d[0]=c=>u.value=c),type:"text",name:"zone-name","data-testid":"name-input","data-test-error-type":Z.value!==null?"invalid-dns-name":void 0,"has-error":Z.value!==null,"error-message":Z.value??void 0,disabled:v.value!==null,onBlur:d[1]||(d[1]=c=>R(u.value))},null,8,["modelValue","data-test-error-type","has-error","error-message","disabled"])]),n(),r(w,{appearance:"primary",class:"mt-4",disabled:W.value,"data-testid":"create-zone-button",onClick:J},{default:l(()=>[p.value?(a(),C(e(X),{key:0,color:e(j),size:e(G)},null,8,["color","size"])):(a(),C(e(_e),{key:1,size:e(G)},null,8,["size"])),n(" "+s(e(o)("zones.form.createZoneButtonLabel")),1)]),_:1},8,["disabled"])])]),n(),v.value!==null?(a(),i(h,{key:0},[t("div",nn,[t("div",on,[t("h2",tn,s(e(o)("zones.form.section.configuration.title")),1),n(),e(o)("zones.form.section.configuration.description")!==" "?(a(),i("p",sn,s(e(o)("zones.form.section.configuration.description")),1)):b("",!0)]),n(),t("div",an,[t("div",rn,[t("div",ln,[t("span",cn,s(e(o)("zones.form.environmentLabel"))+` *
                        `,1),n(),t("div",dn,[r(P,{id:"zone-environment-universal",modelValue:K.value,"onUpdate:modelValue":d[2]||(d[2]=c=>K.value=c),"selected-value":"universal",name:"zone-environment","data-testid":"environment-universal-radio-button"},{default:l(()=>[n(s(e(o)("zones.form.universalLabel")),1)]),_:1},8,["modelValue"]),n(),r(P,{id:"zone-environment-kubernetes",modelValue:K.value,"onUpdate:modelValue":d[3]||(d[3]=c=>K.value=c),"selected-value":"kubernetes",name:"zone-environment","data-testid":"environment-kubernetes-radio-button"},{default:l(()=>[n(s(e(o)("zones.form.kubernetesLabel")),1)]),_:1},8,["modelValue"])])]),n(),K.value==="kubernetes"?(a(),i(h,{key:0},[t("div",un,[t("span",mn,s(e(o)("zones.form.zoneIngressLabel"))+` *
                          `,1),n(),t("div",_n,[r(O,{id:"zone-ingress-enabled",modelValue:L.value,"onUpdate:modelValue":d[4]||(d[4]=c=>L.value=c),"data-testid":"ingress-input-switch"},{label:l(()=>[n(s(e(o)("zones.form.zoneIngressEnabledLabel")),1)]),_:1},8,["modelValue"])])]),n(),t("div",pn,[t("span",fn,s(e(o)("zones.form.zoneEgressLabel"))+` *
                          `,1),n(),t("div",vn,[r(O,{id:"zone-egress-enabled",modelValue:D.value,"onUpdate:modelValue":d[5]||(d[5]=c=>D.value=c),"data-testid":"egress-input-switch"},{label:l(()=>[n(s(e(o)("zones.form.zoneEgressEnabledLabel")),1)]),_:1},8,["modelValue"])])])],64)):b("",!0)])])]),n(),t("div",bn,[t("div",zn,[t("h2",kn,s(e(o)("zones.form.section.connect_zone.title")),1),n(),e(o)("zones.form.section.connect_zone.description")!==" "?(a(),i("p",hn,s(e(o)("zones.form.section.connect_zone.description")),1)):b("",!0)]),n(),t("div",gn,[r(F,{src:"/control-plane/addresses"},{default:l(({data:c})=>[typeof c<"u"?(a(),i(h,{key:0},[K.value==="universal"?(a(),C(Be,{key:0,"zone-name":u.value,token:I.value,"global-kds-address":c.kds},null,8,["zone-name","token","global-kds-address"])):(a(),C(Se,{key:1,"zone-name":u.value,"zone-ingress-enabled":L.value,"zone-egress-enabled":D.value,token:I.value,"base64-encoded-token":Q.value,"global-kds-address":c.kds},null,8,["zone-name","zone-ingress-enabled","zone-egress-enabled","token","base64-encoded-token","global-kds-address"]))],64)):b("",!0)]),_:1})])]),n(),t("div",yn,[t("div",Cn,[t("h2",En,s(e(o)("zones.form.section.scanner.title")),1),n(),e(o)("zones.form.section.scanner.description")!==" "?(a(),i("p",Kn,s(e(o)("zones.form.section.scanner.description")),1)):b("",!0)]),n(),t("div",$n,[r(F,{src:`/zone-cps/online/${u.value}?no-cache=${Date.now()}`,onChange:Y},{default:l(({data:c,error:S})=>[r(se,{"cta-is-hidden":""},{title:l(()=>[S?(a(),i(h,{key:0},[r(e(pe),{display:"inline-block",color:e(fe)},null,8,["color"]),n(),t("h3",null,s(e(o)("zones.form.scan.errorTitle")),1)],64)):typeof c>"u"?(a(),i(h,{key:1},[r(e(X),{"data-testid":"waiting",display:"inline-block",color:e(j)},null,8,["color"]),n(" "+s(e(o)("zones.form.scan.waitTitle")),1)],64)):(a(),i(h,{key:2},[r(e(ve),{"data-testid":"connected",display:"inline-block",color:e(M)},null,8,["color"]),n(" "+s(e(o)("zones.form.scan.completeTitle")),1)],64))]),message:l(()=>[S?(a(),i("p",Vn,s(e(o)("zones.form.scan.errorDescription")),1)):typeof c<"u"?(a(),i(h,{key:1},[t("p",null,s(e(o)("zones.form.scan.completeDescription",{name:u.value})),1),n(),t("p",In,[r(w,{appearance:"primary",to:{name:"zone-cp-detail-view",params:{zone:u.value}}},{default:l(()=>[n(s(e(o)("zones.form.scan.completeButtonLabel",{name:u.value})),1)]),_:1},8,["to"])])],64)):b("",!0)]),_:2},1024)]),_:1},8,["src"])])])],64)):b("",!0)])]),_:1})]),n(),r(re,{"is-visible":V.value,title:e(o)("zones.form.confirm_modal.title"),"data-testid":"confirm-exit-modal",onCanceled:U,onProceed:c=>ce.replace({name:"zone-cp-list-view"})},{"header-content":l(()=>[n(s(e(o)("zones.form.confirm_modal.title")),1)]),"body-content":l(()=>[n(s(e(o)("zones.form.confirm_modal.body")),1)]),"action-buttons":l(()=>[r(w,{appearance:"primary",to:{name:"zone-cp-list-view"},"data-testid":"confirm-exit-button"},{default:l(()=>[n(s(e(o)("zones.form.confirm_modal.action_button")),1)]),_:1})]),_:2},1032,["is-visible","title","onProceed"])]),_:2},1024)]),_:1})}}});const Sn=be(wn,[["__scopeId","data-v-e95cb601"]]);export{Sn as default};
